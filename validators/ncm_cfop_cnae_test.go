package validators

import "testing"

func TestIsValidNCM(t *testing.T) {
	validos := []string{
		"85176211",   // smartphone
		"8517.62.11", // formatado com pontos
		"00000000",   // estruturalmente valido
		"99999999",
	}
	for _, ncm := range validos {
		if !IsValidNCM(ncm) {
			t.Errorf("IsValidNCM(%q) = false, esperado true", ncm)
		}
	}

	invalidos := []struct {
		ncm    string
		motivo string
	}{
		{"", "vazio"},
		{"123", "menos de 8 digitos"},
		{"123456789", "mais de 8 digitos"},
		{"8517621A", "letra no fim"},
		{"abcdefgh", "tudo letra"},
	}
	for _, tc := range invalidos {
		if IsValidNCM(tc.ncm) {
			t.Errorf("IsValidNCM(%q) = true, esperado false (%s)", tc.ncm, tc.motivo)
		}
	}
}

func TestIsValidCFOP(t *testing.T) {
	validos := []string{
		"1101", // entrada UF interna
		"2101", // entrada UF outra
		"3101", // entrada exterior
		"5101", // saida UF interna
		"5.101",
		"6101", // saida UF outra
		"7101", // saida exterior
	}
	for _, cfop := range validos {
		if !IsValidCFOP(cfop) {
			t.Errorf("IsValidCFOP(%q) = false, esperado true", cfop)
		}
	}

	invalidos := []struct {
		cfop   string
		motivo string
	}{
		{"", "vazio"},
		{"101", "menos de 4 digitos"},
		{"51010", "mais de 4 digitos"},
		{"4101", "primeiro digito 4 reservado"},
		{"8101", "primeiro digito 8 nao usado"},
		{"9101", "primeiro digito 9 nao usado"},
		{"0101", "primeiro digito 0 invalido"},
		{"abcd", "tudo letra"},
		{"51A1", "letra no meio"},
	}
	for _, tc := range invalidos {
		if IsValidCFOP(tc.cfop) {
			t.Errorf("IsValidCFOP(%q) = true, esperado false (%s)", tc.cfop, tc.motivo)
		}
	}
}

func TestIsValidCNAE(t *testing.T) {
	validos := []string{
		"6190601",    // New Tecnologia (telecom outros servicos)
		"6190-6/01",  // formato CNAE 2.0
		"61.90-6/01", // com ponto
		"0000000",
		"9999999",
	}
	for _, cnae := range validos {
		if !IsValidCNAE(cnae) {
			t.Errorf("IsValidCNAE(%q) = false, esperado true", cnae)
		}
	}

	invalidos := []struct {
		cnae   string
		motivo string
	}{
		{"", "vazio"},
		{"123456", "menos de 7 digitos"},
		{"12345678", "mais de 7 digitos"},
		{"619060A", "letra no fim"},
		{"abcdefg", "tudo letra"},
	}
	for _, tc := range invalidos {
		if IsValidCNAE(tc.cnae) {
			t.Errorf("IsValidCNAE(%q) = true, esperado false (%s)", tc.cnae, tc.motivo)
		}
	}
}
