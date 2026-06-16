package validators

// IsValidCNAE retorna true quando `s` representa um CNAE (Classificacao
// Nacional de Atividades Economicas) estruturalmente valido: 7 digitos
// numericos. CNAE 2.0 usa formato XXXX-X/XX, totalizando 7 digitos.
//
// CNAE nao tem digito verificador. Validar a existencia da classificacao
// na tabela oficial CONCLA eh Nivel 2, fora de escopo deste pacote.
//
// Aceita entrada formatada com pontos, traco ou barra (`6190-6/01`,
// `61.90-6/01`) — Normalize remove caracteres nao-numericos antes da
// checagem.
func IsValidCNAE(s string) bool {
	digits := Normalize(s)
	if len(digits) != 7 {
		return false
	}
	return isAllDigits(digits)
}
