package aliquotas

import "strings"

// icmsInternoPorUF mapeia a sigla da UF pra aliquota interna media de
// ICMS aplicavel a operacoes telecom + revenda padrao. Valores em
// percentual (ex: 17.0 significa 17%). Fonte: regulamentos estaduais
// vigentes em 2026-06, com referencia cruzada na tabela CONFAZ.
//
// IMPORTANTE: estas sao aliquotas MEDIAS. Setores especificos (energia,
// combustiveis, bebidas frias) podem ter aliquotas diferentes. NCM-
// especifico fica pra evolucao futura quando aparecer demanda concreta.
// FECP/FCP nao esta embutido — caller que precisar trata por fora.
//
// Atualizacao oficial:
//   - bump patch semver com PR documentando fonte da mudanca
//   - se a UF mudar de regra de calculo (ex: monofasico de combustivel),
//     abrir issue antes de mexer
var icmsInternoPorUF = map[string]float64{
	"AC": 17.0,
	"AL": 17.0,
	"AM": 18.0,
	"AP": 18.0,
	"BA": 18.0,
	"CE": 18.0,
	"DF": 18.0,
	"ES": 17.0,
	"GO": 17.0,
	"MA": 18.0,
	"MG": 18.0,
	"MS": 17.0,
	"MT": 17.0,
	"PA": 17.0,
	"PB": 18.0,
	"PE": 18.0,
	"PI": 18.0,
	"PR": 19.0,
	"RJ": 18.0,
	"RN": 18.0,
	"RO": 17.5,
	"RR": 17.0,
	"RS": 17.0,
	"SC": 17.0,
	"SE": 18.0,
	"SP": 18.0,
	"TO": 18.0,
}

// GetICMS retorna a aliquota interna media de ICMS pra UF em percentual
// (ex: 17.0 = 17%). Aceita sigla case-insensitive + trim. Retorna
// (0, false) quando a sigla nao e uma das 27 UFs conhecidas.
//
//	pct, ok := aliquotas.GetICMS("ES")    // 17.0, true
//	pct, ok := aliquotas.GetICMS(" sp ")  // 18.0, true (trim + case)
//	pct, ok := aliquotas.GetICMS("XX")    // 0, false
//
// Caller deve usar (ok==false) pra cair em logica de excecao (ex: regime
// especial). Setores que precisem de aliquota especifica por NCM ainda
// nao sao suportados — sempre passe pelos dados oficiais antes de gravar.
func GetICMS(uf string) (float64, bool) {
	v, ok := icmsInternoPorUF[strings.ToUpper(strings.TrimSpace(uf))]
	return v, ok
}
