package validators

import "testing"

func TestNormalize_RemoveFormatacao(t *testing.T) {
	cases := map[string]string{
		"":                       "",
		"123.456.789-00":         "12345678900",
		"12.345.678/0001-99":     "12345678000199",
		" 12345 6789 ":           "123456789",
		"abc-123-def-456":        "123456",
		"no digits here at all!": "",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, esperado %q", in, got, want)
		}
	}
}

func TestIsValidCPF_validos(t *testing.T) {
	cases := []string{
		"529.982.247-25",
		"52998224725",
		"111.444.777-35",
	}
	for _, cpf := range cases {
		if !IsValidCPF(cpf) {
			t.Errorf("IsValidCPF(%q) = false, esperado true", cpf)
		}
	}
}

func TestIsValidCPF_invalidos(t *testing.T) {
	cases := []struct {
		cpf    string
		motivo string
	}{
		{"", "vazio"},
		{"123", "menos de 11 digitos"},
		{"529.982.247-26", "DV2 errado"},
		{"529.982.247-15", "DV1 errado"},
		{"111.111.111-11", "todos iguais"},
		{"000.000.000-00", "todos zero"},
		{"abc.def.ghi-jk", "sem digitos"},
	}
	for _, tc := range cases {
		if IsValidCPF(tc.cpf) {
			t.Errorf("IsValidCPF(%q) = true, esperado false (%s)", tc.cpf, tc.motivo)
		}
	}
}

func TestIsValidCNPJ_validos(t *testing.T) {
	cases := []string{
		"11.222.333/0001-81",
		"11222333000181",
		"13.332.378/0001-34", // New Tecnologia
	}
	for _, cnpj := range cases {
		if !IsValidCNPJ(cnpj) {
			t.Errorf("IsValidCNPJ(%q) = false, esperado true", cnpj)
		}
	}
}

func TestIsValidCNPJ_invalidos(t *testing.T) {
	cases := []struct {
		cnpj   string
		motivo string
	}{
		{"", "vazio"},
		{"123", "menos de 14 digitos"},
		{"11.222.333/0001-82", "DV2 errado"},
		{"11.222.333/0001-91", "DV1 errado"},
		{"00.000.000/0000-00", "todos zero"},
		{"11.111.111/1111-11", "todos iguais"},
	}
	for _, tc := range cases {
		if IsValidCNPJ(tc.cnpj) {
			t.Errorf("IsValidCNPJ(%q) = true, esperado false (%s)", tc.cnpj, tc.motivo)
		}
	}
}

func TestIsValidDoc_RoteiaPorTipo(t *testing.T) {
	cpfValido := "529.982.247-25"
	cnpjValido := "11.222.333/0001-81"

	cases := []struct {
		doc     string
		docType string
		want    bool
	}{
		{cpfValido, "CPF", true},
		{cpfValido, "cpf", true},
		{cpfValido, " CPF ", true},
		{cpfValido, "CNPJ", false}, // CPF no slot CNPJ
		{cnpjValido, "CNPJ", true},
		{cnpjValido, "cnpj", true},
		{cnpjValido, "CPF", false},
		{cpfValido, "RG", false}, // tipo invalido
		{"", "CPF", false},
	}
	for _, tc := range cases {
		if got := IsValidDoc(tc.doc, tc.docType); got != tc.want {
			t.Errorf("IsValidDoc(%q,%q) = %v, esperado %v", tc.doc, tc.docType, got, tc.want)
		}
	}
}
