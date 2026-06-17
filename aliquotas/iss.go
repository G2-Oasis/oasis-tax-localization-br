package aliquotas

import "strings"

// issPorIBGE mapeia o codigo IBGE de 7 digitos do municipio pra aliquota
// padrao de ISS aplicavel a servicos telecom + revenda. Valores em
// percentual (ex: 5.0 significa 5%).
//
// Cobertura MVP: clientes piloto da G2 + principais capitais. Municipios
// nao cobertos retornam (0, false) — caller deve cair pra Sefin Nacional
// como fallback (NFS-e Nacional cobre todo o territorio).
//
// LIMITE conhecido: ISS eh municipal — Brasil tem 5570 municipios. Cobrir
// todos eh inviavel. Estrategia eh adicionar codigo IBGE conforme novos
// clientes entrarem em homologacao. PR pra cada adicao deve linkar a lei
// municipal vigente.
//
// LC 116/2003 fixa o teto em 5% e o piso em 2% pra ISS. Aliquota efetiva
// varia por servico (item da lista anexa) — esta tabela traz a aliquota
// MAIS COMUM pra servicos telecom (item 1.07 da lista anexa LC 116). Se
// um servico cair em outro item da lista (ex: consultoria), caller deve
// validar separado.
var issPorIBGE = map[string]float64{
	// === clientes piloto G2 ===
	"3201209": 5.0, // Cachoeiro de Itapemirim/ES — New Tecnologia

	// === capitais (cobertura defensiva) ===
	"5300108": 5.0, // Brasilia/DF
	"3550308": 5.0, // Sao Paulo/SP
	"3304557": 5.0, // Rio de Janeiro/RJ
	"3106200": 5.0, // Belo Horizonte/MG
	"4314902": 5.0, // Porto Alegre/RS
	"4106902": 5.0, // Curitiba/PR
	"4205407": 5.0, // Florianopolis/SC
	"2304400": 5.0, // Fortaleza/CE
	"2611606": 5.0, // Recife/PE
	"2927408": 5.0, // Salvador/BA
	"5208707": 5.0, // Goiania/GO
	"1501402": 5.0, // Belem/PA
	"1302603": 5.0, // Manaus/AM
}

// GetISS retorna a aliquota de ISS pro municipio (codigo IBGE de 7
// digitos). Aceita entrada com whitespace. Retorna (0, false) quando o
// municipio nao esta na cobertura — caller deve usar Sefin Nacional
// como fallback.
//
//	pct, ok := aliquotas.GetISS("3201209") // 5.0, true (Cachoeiro)
//	pct, ok := aliquotas.GetISS("9999999") // 0, false (sem cobertura)
//
// Aliquotas refletem a maioria dos servicos telecom (item 1.07 da lista
// anexa da LC 116/2003). Servicos fora do telecom (consultoria,
// engenharia, etc.) podem ter aliquota distinta no mesmo municipio —
// caller precisa validar pela lista de servicos.
func GetISS(ibge string) (float64, bool) {
	v, ok := issPorIBGE[strings.TrimSpace(ibge)]
	return v, ok
}

// IsMunicipioCoberto eh atalho semantico equivalente a `_, ok := GetISS(ibge)`.
// Util quando o caller so precisa decidir se entra no fluxo de NFS-e
// Municipal ou cai pra Sefin Nacional, sem usar a aliquota.
//
//	if aliquotas.IsMunicipioCoberto(prestador.IBGE) {
//	    // roteia pra driver ABRASF municipal
//	} else {
//	    // cai pra Sefin Nacional
//	}
func IsMunicipioCoberto(ibge string) bool {
	_, ok := GetISS(ibge)
	return ok
}
