package validators

// IsValidNCM retorna true quando `s` representa uma NCM (Nomenclatura
// Comum do Mercosul) estruturalmente valida: 8 digitos numericos. NCM
// nao tem digito verificador — validar a existencia da classificacao
// na tabela TIPI eh Nivel 2, fora de escopo deste pacote.
//
// Aceita entrada formatada com pontos (`8517.62.11`) — Normalize remove
// caracteres nao-numericos antes da checagem.
func IsValidNCM(s string) bool {
	digits := Normalize(s)
	if len(digits) != 8 {
		return false
	}
	return isAllDigits(digits)
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
