package reforma

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// exemploRequest replica o JSON de entrada-regime-geral.json do guia oficial
// da Receita, adaptado para um valor simples de teste.
func exemploRequest() RegimeGeralRequest {
	return RegimeGeralRequest{
		ID:              "507f1f77bcf86cd799439011",
		Versao:          "1.0.0",
		DataHoraEmissao: "2026-04-20T12:00:00-03:00",
		Municipio:       4314902,
		UF:              "RS",
		Itens: []ItemInput{
			{
				Numero:      1,
				NCM:         "24021000",
				Quantidade:  1,
				Unidade:     "UN",
				CST:         "550",
				BaseCalculo: 1000.0,
				CClassTrib:  "550020",
			},
		},
	}
}

// parsedResponse espelha o shape oficial da calculadora RFB, que o MockClient
// agora emite. Usado pelos testes para extrair vCBS/vIBS do Raw.
type parsedResponse struct {
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
		} `json:"tribCalc"`
	} `json:"total"`
}

func TestMockCalcularAplicaAliquotasTeste2026(t *testing.T) {
	client := NewMockClient()

	resp, err := client.Calcular(context.Background(), exemploRequest())
	if err != nil {
		t.Fatalf("calcular: %v", err)
	}

	// Item com baseCalculo=1000 → CBS=9.00 (0,9%) e IBS=1.00 (0,1%).
	// Valores vem como strings formatadas (shape oficial da RFB).
	var parsed parsedResponse
	if err := json.Unmarshal(resp.Raw, &parsed); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(parsed.Objetos) != 1 {
		t.Fatalf("esperado 1 objeto, veio %d", len(parsed.Objetos))
	}
	if got := parsed.Objetos[0].TribCalc.IBSCBS.GIBSCBS.GCBS.VCBS; got != "9.00" {
		t.Errorf("vCBS: esperado \"9.00\", veio %q", got)
	}
	if got := parsed.Objetos[0].TribCalc.IBSCBS.GIBSCBS.VIBS; got != "1.00" {
		t.Errorf("vIBS: esperado \"1.00\", veio %q", got)
	}
	if got := parsed.Total.TribCalc.IBSCBSTot.GCBS.VCBS; got != "9.00" {
		t.Errorf("total vCBS: esperado \"9.00\", veio %q", got)
	}
}

func TestMockCalcularAliquotasCustomizadas(t *testing.T) {
	client := NewMockClient()
	client.AliquotaCBS = 0.08 // 8% (aliquota cheia hipotetica)
	client.AliquotaIBS = 0.18 // 18%

	resp, _ := client.Calcular(context.Background(), exemploRequest())

	var parsed parsedResponse
	_ = json.Unmarshal(resp.Raw, &parsed)

	if got := parsed.Objetos[0].TribCalc.IBSCBS.GIBSCBS.GCBS.VCBS; got != "80.00" {
		t.Errorf("vCBS customizado: esperado \"80.00\", veio %q", got)
	}
	if got := parsed.Objetos[0].TribCalc.IBSCBS.GIBSCBS.VIBS; got != "180.00" {
		t.Errorf("vIBS customizado: esperado \"180.00\", veio %q", got)
	}
}

func TestMockCalcularInjetaErro(t *testing.T) {
	client := NewMockClient()
	client.ForceCalcularError = errors.New("calculadora offline")

	_, err := client.Calcular(context.Background(), exemploRequest())
	if err == nil {
		t.Fatal("esperado erro injetado")
	}
	if !strings.Contains(err.Error(), "offline") {
		t.Errorf("mensagem de erro inesperada: %v", err)
	}
}

func TestMockCalcularRespeitaContextDeadline(t *testing.T) {
	client := NewMockClient()
	client.Delay = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Calcular(ctx, exemploRequest())
	if err == nil {
		t.Fatal("esperado context deadline exceeded")
	}
}

func TestMockGerarXMLContemGruposRTC(t *testing.T) {
	client := NewMockClient()
	calc, _ := client.Calcular(context.Background(), exemploRequest())

	xml, err := client.GerarXML(context.Background(), "nfe", calc)
	if err != nil {
		t.Fatalf("gerar xml: %v", err)
	}
	s := string(xml)

	// Mock emite no shape oficial da RFB (<infNFe>/<det>/<imposto>/<IBSCBS>).
	// O item de exemplo usa CST 000 (sem IS), entao IS/ISTot nao aparecem —
	// a calculadora real omite esses grupos quando nao se aplicam.
	for _, tag := range []string{"<infNFe", "<det nItem=\"1\">", "<imposto>", "<IBSCBS>", "<total>", "<IBSCBSTot>"} {
		if !strings.Contains(s, tag) {
			t.Errorf("XML deveria conter %q, nao contem:\n%s", tag, s)
		}
	}
}

// Com IS configurado > 0, os grupos <IS> e <ISTot> aparecem (espelhando
// o comportamento da calculadora real que so os emite quando aplicaveis).
func TestMockGerarXMLInclueIS_QuandoISAplica(t *testing.T) {
	client := NewMockClient()
	client.AliquotaIS = 0.1 // 10%
	req := exemploRequest()
	req.Itens[0].ImpostoSeletivo = &ImpostoSeletivo{BaseCalculo: 500.0}

	calc, _ := client.Calcular(context.Background(), req)
	xml, err := client.GerarXML(context.Background(), "nfe", calc)
	if err != nil {
		t.Fatalf("gerar xml: %v", err)
	}
	s := string(xml)
	for _, tag := range []string{"<IS>", "<ISTot>"} {
		if !strings.Contains(s, tag) {
			t.Errorf("XML deveria conter %q com IS aplicavel:\n%s", tag, s)
		}
	}
}

func TestMockValidarXMLValidoPorPadrao(t *testing.T) {
	client := NewMockClient()

	resp, err := client.ValidarXML(context.Background(), "nfe", "grupo", []byte("<xml/>"))
	if err != nil {
		t.Fatalf("validar: %v", err)
	}
	if !resp.Valido {
		t.Fatalf("esperado valido, veio invalido: %s", resp.Mensagem)
	}
}

func TestMockValidarXMLInvalidoForcado(t *testing.T) {
	client := NewMockClient()
	client.ForceInvalido = true
	client.MensagemInvalido = "tag <vIBS> obrigatoria"

	resp, _ := client.ValidarXML(context.Background(), "nfe", "grupo", []byte("<xml/>"))
	if resp.Valido {
		t.Fatal("esperado invalido")
	}
	if !strings.Contains(resp.Mensagem, "vIBS") {
		t.Errorf("mensagem inesperada: %s", resp.Mensagem)
	}
}

// Garantia estatica de que MockClient implementa a interface Client.
var _ Client = (*MockClient)(nil)
