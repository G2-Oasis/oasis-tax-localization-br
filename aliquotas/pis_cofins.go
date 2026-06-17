package aliquotas

import "strings"

// Regime tributario federal — chaves canonicas usadas em GetPISCofins.
const (
	// RegimeSimplesNacional cobre microempresa + empresa de pequeno porte
	// optantes do Simples (LC 123/2006). PIS/COFINS estao embutidos no
	// DAS unico — a contribuinte nao recolhe separado.
	RegimeSimplesNacional = "SIMPLES_NACIONAL"

	// RegimeLucroReal cobre empresas em apuracao nao-cumulativa
	// (Leis 10.637/2002 e 10.833/2003). Aliquotas oficiais 1.65% / 7.60%.
	RegimeLucroReal = "LUCRO_REAL"

	// RegimeLucroPresumido cobre empresas em apuracao cumulativa.
	// Aliquotas oficiais 0.65% / 3.0%.
	RegimeLucroPresumido = "LUCRO_PRESUMIDO"
)

// PISCofins agrupa as duas aliquotas vigentes pro regime tributario.
// Valores em percentual (ex: 1.65 significa 1,65%).
type PISCofins struct {
	PIS    float64
	COFINS float64
}

// pisCofinsPorRegime mapeia o regime tributario federal pras aliquotas
// vigentes. Fonte: legislacao federal (Leis 9.715/1998, 10.637/2002,
// 10.833/2003) + tabelas SEFAZ.
var pisCofinsPorRegime = map[string]PISCofins{
	RegimeSimplesNacional: {PIS: 0.0, COFINS: 0.0}, // embutidos no DAS
	RegimeLucroReal:       {PIS: 1.65, COFINS: 7.60},
	RegimeLucroPresumido:  {PIS: 0.65, COFINS: 3.0},
}

// GetPISCofins retorna PIS e COFINS pro regime tributario. Aceita chave
// case-insensitive + trim. Retorna (zero, false) quando o regime nao eh
// reconhecido — use as constantes Regime* pra evitar typo.
//
//	pc, ok := aliquotas.GetPISCofins(aliquotas.RegimeLucroReal)
//	// pc.PIS == 1.65, pc.COFINS == 7.60, ok == true
//
// Simples Nacional retorna {0, 0, true}: as contribuicoes estao
// embutidas no DAS unico, entao a aliquota efetiva pra fins de calculo
// item-a-item eh zero — caller pode tratar o caso especialmente quando
// precisar exibir "n/a" na UI.
func GetPISCofins(regime string) (PISCofins, bool) {
	v, ok := pisCofinsPorRegime[strings.ToUpper(strings.TrimSpace(regime))]
	return v, ok
}
