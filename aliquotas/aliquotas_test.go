package aliquotas

import "testing"

func TestGetICMS_Cobertura27UFs(t *testing.T) {
	if got := len(icmsInternoPorUF); got != 27 {
		t.Errorf("icmsInternoPorUF deveria ter 27 entradas (26 estados + DF), got %d", got)
	}
}

func TestGetICMS_AmostraConhecida(t *testing.T) {
	cases := map[string]float64{
		"ES": 17.0, // New Tecnologia
		"SP": 18.0,
		"RJ": 18.0,
		"PR": 19.0,
		"RS": 17.0,
		"DF": 18.0,
	}
	for uf, expected := range cases {
		got, ok := GetICMS(uf)
		if !ok || got != expected {
			t.Errorf("GetICMS(%q) = %v,%v; esperado %v,true", uf, got, ok, expected)
		}
	}
}

func TestGetICMS_CaseInsensitiveETrim(t *testing.T) {
	cases := []string{"es", " ES ", "Es", "eS", "\tES\n"}
	for _, in := range cases {
		got, ok := GetICMS(in)
		if !ok || got != 17.0 {
			t.Errorf("GetICMS(%q) = %v,%v; esperado 17.0,true", in, got, ok)
		}
	}
}

func TestGetICMS_UFInvalida(t *testing.T) {
	cases := []string{"", "XX", "ZZ", "BR", "USA", "12"}
	for _, in := range cases {
		if _, ok := GetICMS(in); ok {
			t.Errorf("GetICMS(%q) deveria retornar ok=false", in)
		}
	}
}

func TestGetICMS_LimitesLegais(t *testing.T) {
	// LC 87/1996 + Resolucao SF 22/89: aliquota interna minima 7%, maxima
	// pode chegar a 25% em setores especificos. Garantir que nenhuma
	// entrada do MVP escapou desses limites por typo.
	for uf, pct := range icmsInternoPorUF {
		if pct < 7.0 || pct > 25.0 {
			t.Errorf("ICMS de %s fora do dominio legal [7%%-25%%]: %.2f%%", uf, pct)
		}
	}
}

func TestGetPISCofins_RegimesConhecidos(t *testing.T) {
	cases := map[string]PISCofins{
		RegimeSimplesNacional: {PIS: 0.0, COFINS: 0.0},
		RegimeLucroReal:       {PIS: 1.65, COFINS: 7.60},
		RegimeLucroPresumido:  {PIS: 0.65, COFINS: 3.0},
	}
	for regime, expected := range cases {
		got, ok := GetPISCofins(regime)
		if !ok {
			t.Errorf("GetPISCofins(%q) ok=false", regime)
			continue
		}
		if got != expected {
			t.Errorf("GetPISCofins(%q) = %+v; esperado %+v", regime, got, expected)
		}
	}
}

func TestGetPISCofins_CaseInsensitiveETrim(t *testing.T) {
	cases := []string{
		"LUCRO_REAL",
		"lucro_real",
		" Lucro_Real ",
		"\tLUCRO_REAL\n",
	}
	for _, regime := range cases {
		got, ok := GetPISCofins(regime)
		if !ok || got.PIS != 1.65 || got.COFINS != 7.60 {
			t.Errorf("GetPISCofins(%q) = %+v,%v; esperado {1.65, 7.60},true", regime, got, ok)
		}
	}
}

func TestGetPISCofins_RegimeDesconhecido(t *testing.T) {
	cases := []string{"", "MEI", "ARBITRADO", "XX"}
	for _, regime := range cases {
		if _, ok := GetPISCofins(regime); ok {
			t.Errorf("GetPISCofins(%q) deveria retornar ok=false", regime)
		}
	}
}

func TestGetISS_ClientesPiloto(t *testing.T) {
	cases := map[string]float64{
		"3201209": 5.0, // Cachoeiro de Itapemirim/ES — New Tec
		"3550308": 5.0, // Sao Paulo
		"3304557": 5.0, // Rio de Janeiro
	}
	for ibge, expected := range cases {
		got, ok := GetISS(ibge)
		if !ok || got != expected {
			t.Errorf("GetISS(%q) = %v,%v; esperado %v,true", ibge, got, ok, expected)
		}
	}
}

func TestGetISS_MunicipioForaCobertura(t *testing.T) {
	cases := []string{"", "9999999", "1234567", "abc1234", "3201208"}
	for _, ibge := range cases {
		if _, ok := GetISS(ibge); ok {
			t.Errorf("GetISS(%q) deveria retornar ok=false (fora da cobertura MVP)", ibge)
		}
	}
}

func TestGetISS_LimitesLegais(t *testing.T) {
	// LC 116/2003 fixa teto de 5% e piso de 2% pra ISS.
	for ibge, pct := range issPorIBGE {
		if pct < 2.0 || pct > 5.0 {
			t.Errorf("ISS de %s fora do dominio legal [2%%-5%%]: %.2f%%", ibge, pct)
		}
	}
}

func TestIsMunicipioCoberto(t *testing.T) {
	if !IsMunicipioCoberto("3201209") {
		t.Error("Cachoeiro de Itapemirim deveria estar coberto (cliente piloto)")
	}
	if IsMunicipioCoberto("9999999") {
		t.Error("IBGE inexistente nao deveria estar coberto")
	}
	if IsMunicipioCoberto("") {
		t.Error("string vazia nao deveria estar coberta")
	}
}
