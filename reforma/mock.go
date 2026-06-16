package reforma

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Aliquotas oficiais do periodo de teste da Reforma Tributaria (2026).
// A partir de 2027, as aliquotas crescem progressivamente conforme LC 214/2025.
// Mantemos as 2026 como default do MockClient por serem as vigentes hoje.
const (
	AliquotaCBSTeste2026 = 0.009 // 0,9%
	AliquotaIBSTeste2026 = 0.001 // 0,1%
	AliquotaISDefault    = 0.0   // sem IS por padrao (item nao sujeito)
)

// MockClient simula a calculadora para dev e testes — sem rede, sem Docker.
//
// Comportamento default: aplica as aliquotas oficiais de teste de 2026
// sobre a baseCalculo de cada item.
//
// Cenarios de falha podem ser configurados via os campos ForceCalcularError,
// ForceGerarXMLError e ForceInvalido — padrao que espelha o MockClient da
// integracao SEFAZ usado no Bloco 7.
type MockClient struct {
	mu sync.RWMutex

	// Aliquotas configuraveis (se zero, usa os defaults 2026).
	AliquotaCBS float64
	AliquotaIBS float64
	AliquotaIS  float64

	// Delay simulado antes de responder (respeita context deadline).
	Delay time.Duration

	// Injecao de erros para testar caminhos de falha.
	ForceCalcularError   error
	ForceGerarXMLError   error
	ForceValidarError    error
	ForceInvalido        bool
	MensagemInvalido     string
}

func NewMockClient() *MockClient {
	return &MockClient{
		AliquotaCBS: AliquotaCBSTeste2026,
		AliquotaIBS: AliquotaIBSTeste2026,
		AliquotaIS:  AliquotaISDefault,
	}
}

// Calcular simula o calculo de CBS/IBS/IS aplicando aliquotas fixas sobre a
// baseCalculo de cada item. O JSON de resposta preserva a estrutura de
// entrada e adiciona um bloco `tributos` por item — suficiente para o
// restante do fluxo repassar a GerarXML.
func (m *MockClient) Calcular(ctx context.Context, req RegimeGeralRequest) (RegimeGeralResponse, error) {
	if err := m.simulateDelay(ctx); err != nil {
		return RegimeGeralResponse{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ForceCalcularError != nil {
		return RegimeGeralResponse{}, m.ForceCalcularError
	}

	aliquotaCBS := m.AliquotaCBS
	if aliquotaCBS == 0 {
		aliquotaCBS = AliquotaCBSTeste2026
	}
	aliquotaIBS := m.AliquotaIBS
	if aliquotaIBS == 0 {
		aliquotaIBS = AliquotaIBSTeste2026
	}

	// Resposta espelha o shape oficial da calculadora RFB (LC 214/2025).
	// Manter mock e sidecar emitindo o mesmo formato garante que o parser
	// se comporte igual em dev e producao.
	//
	// Simplificacoes vs. resposta real:
	//   - memoriaCalculo omitida
	//   - creditos presumidos omitidos (vCredPres, vCredPresCondSus)
	//   - parcela municipal (gIBSMun) zerada — toda a aliquota IBS fica em UF
	totalCBS, totalIBS, totalIS, totalBC := 0.0, 0.0, 0.0, 0.0
	objetos := make([]map[string]any, 0, len(req.Itens))
	for _, item := range req.Itens {
		vCBS := round2(item.BaseCalculo * aliquotaCBS)
		vIBS := round2(item.BaseCalculo * aliquotaIBS)
		vIS := 0.0
		if item.ImpostoSeletivo != nil {
			vIS = round2(item.ImpostoSeletivo.BaseCalculo * m.AliquotaIS)
		}
		totalCBS += vCBS
		totalIBS += vIBS
		totalIS += vIS
		totalBC += item.BaseCalculo

		tribCalc := map[string]any{
			"IBSCBS": map[string]any{
				"CST":        item.CST,
				"cClassTrib": item.CClassTrib,
				"gIBSCBS": map[string]any{
					"vBC": formatMoney(item.BaseCalculo),
					"gIBSUF": map[string]any{
						"pIBSUF": formatRate(aliquotaIBS * 100),
						"vIBSUF": formatMoney(vIBS),
					},
					"gIBSMun": map[string]any{
						"pIBSMun": "0.00",
						"vIBSMun": "0.00",
					},
					"vIBS": formatMoney(vIBS),
					"gCBS": map[string]any{
						"pCBS": formatRate(aliquotaCBS * 100),
						"vCBS": formatMoney(vCBS),
					},
				},
			},
		}
		// IS so aparece no response quando o item esta sujeito ao imposto
		// seletivo — espelha o comportamento da calculadora real.
		if vIS != 0 {
			tribCalc["IS"] = map[string]any{"vIS": formatMoney(vIS)}
		}

		objetos = append(objetos, map[string]any{
			"nObj":     item.Numero,
			"tribCalc": tribCalc,
		})
	}

	totalTribCalc := map[string]any{
		"IBSCBSTot": map[string]any{
			"vBCIBSCBS": formatMoney(totalBC),
			"gIBS":      map[string]any{"vIBS": formatMoney(totalIBS)},
			"gCBS":      map[string]any{"vCBS": formatMoney(totalCBS)},
		},
	}
	if totalIS != 0 {
		totalTribCalc["ISTot"] = map[string]any{"vIS": formatMoney(totalIS)}
	}

	payload := map[string]any{
		"id":              req.ID,
		"versao":          req.Versao,
		"dataHoraEmissao": req.DataHoraEmissao,
		"municipio":       req.Municipio,
		"uf":              req.UF,
		"objetos":         objetos,
		"total":           map[string]any{"tribCalc": totalTribCalc},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return RegimeGeralResponse{}, fmt.Errorf("mock: marshal response: %w", err)
	}
	return RegimeGeralResponse{Raw: raw}, nil
}

// GerarXML produz um XML plausivel com os 4 grupos RTC esperados.
// E representativo, nao e validavel contra XSD — propositalmente simples
// para servir apenas ao fluxo de desenvolvimento.
func (m *MockClient) GerarXML(ctx context.Context, tipo string, calculo RegimeGeralResponse) ([]byte, error) {
	if err := m.simulateDelay(ctx); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ForceGerarXMLError != nil {
		return nil, m.ForceGerarXMLError
	}

	// Parse do JSON retornado por Calcular no shape oficial da RFB para
	// remontar o XML RTC. Campos ausentes viram zero.
	var parsed struct {
		Objetos []struct {
			NObj     int `json:"nObj"`
			TribCalc struct {
				IBSCBS struct {
					GIBSCBS struct {
						GCBS struct {
							VCBS string `json:"vCBS"`
						} `json:"gCBS"`
						VIBS string `json:"vIBS"`
					} `json:"gIBSCBS"`
				} `json:"IBSCBS"`
				IS struct {
					VIS string `json:"vIS"`
				} `json:"IS"`
			} `json:"tribCalc"`
		} `json:"objetos"`
		Total struct {
			TribCalc struct {
				IBSCBSTot struct {
					GCBS struct {
						VCBS string `json:"vCBS"`
					} `json:"gCBS"`
					GIBS struct {
						VIBS string `json:"vIBS"`
					} `json:"gIBS"`
				} `json:"IBSCBSTot"`
				ISTot struct {
					VIS string `json:"vIS"`
				} `json:"ISTot"`
			} `json:"tribCalc"`
		} `json:"total"`
	}
	_ = json.Unmarshal(calculo.Raw, &parsed)

	// Gera XML no shape real da calculadora RFB (capturado em 2026-04-23):
	// <infNFe xmlns="http://www.portalfiscal.inf.br/nfe">
	//   <det nItem="N"><imposto><IBSCBS>...</IBSCBS></imposto></det>
	//   <total><IBSCBSTot>...</IBSCBSTot></total>
	// </infNFe>
	//
	// O elemento raiz depende do `tipo` (nfe, nfce, cte, bpe...). Para
	// nfcom usaria-se <infNFCom>. Aqui replicamos o comportamento da RFB
	// que hoje so oferece layouts de NF-e; o ParseRTC nao depende do
	// nome do root, entao mantemos <infNFe> como default seguro.
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<infNFe xmlns="http://www.portalfiscal.inf.br/nfe">` + "\n")
	for _, o := range parsed.Objetos {
		vCBS := parseMockMoney(o.TribCalc.IBSCBS.GIBSCBS.GCBS.VCBS)
		vIBS := parseMockMoney(o.TribCalc.IBSCBS.GIBSCBS.VIBS)
		vIS := parseMockMoney(o.TribCalc.IS.VIS)
		fmt.Fprintf(&b, `  <det nItem="%d">`+"\n", o.NObj)
		b.WriteString("    <imposto>\n")
		if vIS != 0 {
			fmt.Fprintf(&b, `      <IS><vIS>%.2f</vIS></IS>`+"\n", vIS)
		}
		fmt.Fprintf(&b, `      <IBSCBS><vCBS>%.2f</vCBS><vIBS>%.2f</vIBS></IBSCBS>`+"\n", vCBS, vIBS)
		b.WriteString("    </imposto>\n")
		b.WriteString("  </det>\n")
	}
	totalCBS := parseMockMoney(parsed.Total.TribCalc.IBSCBSTot.GCBS.VCBS)
	totalIBS := parseMockMoney(parsed.Total.TribCalc.IBSCBSTot.GIBS.VIBS)
	totalIS := parseMockMoney(parsed.Total.TribCalc.ISTot.VIS)
	b.WriteString("  <total>\n")
	if totalIS != 0 {
		fmt.Fprintf(&b, `    <ISTot><vIS>%.2f</vIS></ISTot>`+"\n", totalIS)
	}
	fmt.Fprintf(&b, `    <IBSCBSTot><vCBS>%.2f</vCBS><vIBS>%.2f</vIBS></IBSCBSTot>`+"\n", totalCBS, totalIBS)
	b.WriteString("  </total>\n")
	b.WriteString("</infNFe>\n")
	return []byte(b.String()), nil
}

// parseMockMoney converte string "9.00" em float — helper local do mock
// para nao precisar expor o parser do engine. Strings invalidas viram 0.
func parseMockMoney(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// ValidarXML retorna valido por padrao. Configure ForceInvalido para cenarios
// de rejeicao estrutural.
func (m *MockClient) ValidarXML(ctx context.Context, tipo, subtipo string, xml []byte) (ValidarXMLResponse, error) {
	if err := m.simulateDelay(ctx); err != nil {
		return ValidarXMLResponse{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ForceValidarError != nil {
		return ValidarXMLResponse{}, m.ForceValidarError
	}
	if m.ForceInvalido {
		msg := m.MensagemInvalido
		if msg == "" {
			msg = "XML invalido (forcado pelo mock)"
		}
		return ValidarXMLResponse{Valido: false, Mensagem: msg}, nil
	}
	return ValidarXMLResponse{Valido: true, Mensagem: "XML valido"}, nil
}

func (m *MockClient) simulateDelay(ctx context.Context) error {
	if m.Delay <= 0 {
		return nil
	}
	select {
	case <-time.After(m.Delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// formatMoney serializa valores monetarios como a calculadora real: string
// com 2 casas decimais. Ex.: 9.0 -> "9.00".
func formatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// formatRate serializa aliquotas como a calculadora real: string com 2 casas
// decimais. Ex.: 0.9 -> "0.90".
func formatRate(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
