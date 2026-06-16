package validators

import "testing"

func TestUFToIBGECode_Cobre27UFs(t *testing.T) {
	if got := len(ufIBGECodes); got != 27 {
		t.Errorf("ufIBGECodes deveria ter 27 entradas (26 estados + DF), got %d", got)
	}
}

func TestUFToIBGECode_AmostraConhecida(t *testing.T) {
	cases := map[string]string{
		"AC": "12", "ES": "32", "RJ": "33", "RS": "43", "SP": "35",
		"DF": "53", "MG": "31", "BA": "29", "TO": "17", "RR": "14",
	}
	for uf, expected := range cases {
		got, ok := UFToIBGECode(uf)
		if !ok || got != expected {
			t.Errorf("UFToIBGECode(%q) = %q,%v; esperado %q,true", uf, got, ok, expected)
		}
	}
}

func TestUFToIBGECode_CaseInsensitiveETrim(t *testing.T) {
	cases := []string{"es", " ES ", "Es", "eS", "\tES\n"}
	for _, in := range cases {
		got, ok := UFToIBGECode(in)
		if !ok || got != "32" {
			t.Errorf("UFToIBGECode(%q) = %q,%v; esperado 32,true", in, got, ok)
		}
	}
}

func TestUFToIBGECode_Invalida(t *testing.T) {
	cases := []string{"", "  ", "XX", "ZZ", "BR", "USA", "12"}
	for _, in := range cases {
		if _, ok := UFToIBGECode(in); ok {
			t.Errorf("UFToIBGECode(%q) deveria retornar ok=false", in)
		}
	}
}

func TestIBGECodeToUF_AmostraConhecida(t *testing.T) {
	cases := map[string]string{
		"32": "ES",
		"33": "RJ",
		"43": "RS",
		"35": "SP",
		"53": "DF",
		"17": "TO",
		"14": "RR",
	}
	for cUF, expected := range cases {
		got, ok := IBGECodeToUF(cUF)
		if !ok || got != expected {
			t.Errorf("IBGECodeToUF(%q) = %q,%v; esperado %q,true", cUF, got, ok, expected)
		}
	}
}

func TestIBGECodeToUF_Trim(t *testing.T) {
	cases := []string{"32", " 32", "32 ", " 32 ", "\t32\n"}
	for _, in := range cases {
		got, ok := IBGECodeToUF(in)
		if !ok || got != "ES" {
			t.Errorf("IBGECodeToUF(%q) = %q,%v; esperado ES,true", in, got, ok)
		}
	}
}

func TestIBGECodeToUF_Invalido(t *testing.T) {
	cases := []string{"", "99", "00", "X1", "ES", "123"}
	for _, in := range cases {
		if _, ok := IBGECodeToUF(in); ok {
			t.Errorf("IBGECodeToUF(%q) deveria retornar ok=false", in)
		}
	}
}

func TestRoundTrip_UFToIBGE_IBGEToUF(t *testing.T) {
	// Toda sigla -> cUF -> sigla deve voltar igual
	for sigla := range ufIBGECodes {
		cUF, ok := UFToIBGECode(sigla)
		if !ok {
			t.Errorf("UFToIBGECode(%q) ok=false", sigla)
			continue
		}
		back, ok := IBGECodeToUF(cUF)
		if !ok || back != sigla {
			t.Errorf("round-trip %q -> %q -> %q,%v; esperado %q", sigla, cUF, back, ok, sigla)
		}
	}
}
