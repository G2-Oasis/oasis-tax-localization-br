// Package validators reune validacoes estruturais de codigos fiscais
// brasileiros — NCM, CFOP, CNAE, CNPJ, CPF.
//
// Validacoes sao Nivel 1 (estrutural): tamanho, formato, e — quando
// aplicavel — digitos verificadores. Nivel 2 (existencia no cadastro
// oficial via RFB/Serpro/BrasilAPI) eh fora de escopo.
package validators

import "strings"

// Normalize remove caracteres de formatacao comuns ('.', '-', '/', ' ')
// e devolve apenas os digitos. Util pra aceitar entradas como
// "12.345.678/0001-99" ou "123.456.789-00" sem onerar o caller.
func Normalize(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsValidCPF retorna true quando `s` representa um CPF cujos dois digitos
// verificadores conferem pelo algoritmo modulo 11. Aceita entrada formatada
// (normaliza internamente). Rejeita strings com todos os digitos iguais
// (000.000.000-00 ate 999.999.999-99) — sao matematicamente validas mas
// reservadas pela RFB e nunca emitidas.
func IsValidCPF(s string) bool {
	digits := Normalize(s)
	if len(digits) != 11 {
		return false
	}
	if allSameDigit(digits) {
		return false
	}

	// DV1: pesos 10..2 sobre os 9 primeiros digitos
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	dv1 := 11 - (sum % 11)
	if dv1 >= 10 {
		dv1 = 0
	}
	if dv1 != int(digits[9]-'0') {
		return false
	}

	// DV2: pesos 11..2 sobre os 10 primeiros digitos
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	dv2 := 11 - (sum % 11)
	if dv2 >= 10 {
		dv2 = 0
	}
	return dv2 == int(digits[10]-'0')
}

// IsValidCNPJ retorna true quando `s` representa um CNPJ cujos dois digitos
// verificadores conferem pelo algoritmo modulo 11 com pesos especificos.
// Aceita entrada formatada. Rejeita strings com todos os digitos iguais.
func IsValidCNPJ(s string) bool {
	digits := Normalize(s)
	if len(digits) != 14 {
		return false
	}
	if allSameDigit(digits) {
		return false
	}

	// DV1: pesos 5,4,3,2,9,8,7,6,5,4,3,2 sobre os 12 primeiros digitos
	weights1 := [12]int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(digits[i]-'0') * weights1[i]
	}
	dv1 := 11 - (sum % 11)
	if dv1 >= 10 {
		dv1 = 0
	}
	if dv1 != int(digits[12]-'0') {
		return false
	}

	// DV2: pesos 6,5,4,3,2,9,8,7,6,5,4,3,2 sobre os 13 primeiros digitos
	weights2 := [13]int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		sum += int(digits[i]-'0') * weights2[i]
	}
	dv2 := 11 - (sum % 11)
	if dv2 >= 10 {
		dv2 = 0
	}
	return dv2 == int(digits[13]-'0')
}

// IsValidDoc roteia para IsValidCPF ou IsValidCNPJ conforme `docType`
// ("CPF" ou "CNPJ", case-insensitive). Qualquer outro valor devolve false.
// Util pro RecipientProfile que ja carrega o tipo separadamente.
func IsValidDoc(doc, docType string) bool {
	switch strings.ToUpper(strings.TrimSpace(docType)) {
	case "CPF":
		return IsValidCPF(doc)
	case "CNPJ":
		return IsValidCNPJ(doc)
	}
	return false
}

func allSameDigit(digits string) bool {
	first := digits[0]
	for i := 1; i < len(digits); i++ {
		if digits[i] != first {
			return false
		}
	}
	return true
}
