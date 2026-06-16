package validators

// IsValidCFOP retorna true quando `s` representa um CFOP (Codigo Fiscal de
// Operacoes e Prestacoes) estruturalmente valido: 4 digitos numericos
// comecando com 1, 2, 3, 5, 6 ou 7. CFOP nao tem digito verificador.
//
// Classes:
//   - 1xxx: entrada UF interna
//   - 2xxx: entrada UF outra
//   - 3xxx: entrada exterior
//   - 5xxx: saida UF interna
//   - 6xxx: saida UF outra
//   - 7xxx: saida exterior
//
// 4xxx e 8xxx-9xxx nao sao usados na tabela oficial. Validar se o CFOP
// existe na tabela oficial completa eh Nivel 2, fora de escopo.
//
// Aceita entrada formatada com ponto (`5.101`) — Normalize remove
// caracteres nao-numericos antes da checagem.
func IsValidCFOP(s string) bool {
	digits := Normalize(s)
	if len(digits) != 4 {
		return false
	}
	if !isAllDigits(digits) {
		return false
	}
	switch digits[0] {
	case '1', '2', '3', '5', '6', '7':
		return true
	}
	return false
}
