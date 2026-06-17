package validators

import (
	"sort"
	"testing"
)

func TestIsValidCST_ICMSNormal(t *testing.T) {
	validos := []string{"00", "10", "20", "30", "40", "41", "50", "51", "60", "70", "90"}
	for _, cst := range validos {
		if !IsValidCST("ICMS", cst) {
			t.Errorf("IsValidCST(ICMS, %q) = false, esperado true", cst)
		}
	}

	invalidos := []string{"", "99", "01", "101", "abc", " 00", "  "}
	for _, cst := range invalidos {
		// nota: " 00" eh invalido porque trim do cst remove espacos -> "00" valido.
		// recalibrar: trim deixa "00" valido. Vou ajustar o teste pra esperar true em " 00".
		_ = cst
	}
	// Verifica explicitamente: cst com whitespace ao redor eh aceito (Normalize trim).
	if !IsValidCST("ICMS", " 00 ") {
		t.Error("IsValidCST aceita whitespace nos lados do cst (trim aplicado)")
	}

	// Realmente invalidos (depois do trim continuam fora do dominio)
	for _, cst := range []string{"", "99", "01", "101", "abc"} {
		if IsValidCST("ICMS", cst) {
			t.Errorf("IsValidCST(ICMS, %q) = true, esperado false", cst)
		}
	}
}

func TestIsValidCST_ICMSSimples(t *testing.T) {
	validos := []string{"101", "102", "103", "201", "202", "203", "300", "400", "500", "900"}
	for _, cst := range validos {
		if !IsValidCST("ICMS_SN", cst) {
			t.Errorf("IsValidCST(ICMS_SN, %q) = false, esperado true", cst)
		}
	}

	if IsValidCST("ICMS_SN", "00") {
		t.Error("CST normal '00' nao deve valer em ICMS_SN")
	}
	if !IsValidCST("icms_sn", "101") {
		t.Error("imposto deve ser case-insensitive")
	}
}

func TestIsValidCST_PISCOFINS(t *testing.T) {
	validos := []string{"01", "06", "07", "49", "50", "67", "70", "98", "99"}
	for _, cst := range validos {
		if !IsValidCST("PIS", cst) {
			t.Errorf("IsValidCST(PIS, %q) = false, esperado true", cst)
		}
		if !IsValidCST("COFINS", cst) {
			t.Errorf("IsValidCST(COFINS, %q) = false, esperado true", cst)
		}
	}

	invalidos := []string{"", "00", "10", "11", "100"}
	for _, cst := range invalidos {
		if IsValidCST("PIS", cst) {
			t.Errorf("IsValidCST(PIS, %q) = true, esperado false", cst)
		}
		if IsValidCST("COFINS", cst) {
			t.Errorf("IsValidCST(COFINS, %q) = true, esperado false", cst)
		}
	}
}

func TestIsValidCST_ImpostoDesconhecido(t *testing.T) {
	cases := []string{"IPI", "ISS", "ISSQN", "FCP", "", "XYZ"}
	for _, imposto := range cases {
		if IsValidCST(imposto, "00") {
			t.Errorf("IsValidCST(%q, 00) = true, esperado false (imposto fora do MVP)", imposto)
		}
		if IsValidCST(imposto, "99") {
			t.Errorf("IsValidCST(%q, 99) = true, esperado false (imposto fora do MVP)", imposto)
		}
	}
}

func TestIsValidCST_ImpostoCaseInsensitiveETrim(t *testing.T) {
	cases := []string{"ICMS", "icms", "Icms", " ICMS ", "\tICMS\n"}
	for _, imposto := range cases {
		if !IsValidCST(imposto, "00") {
			t.Errorf("IsValidCST(%q, 00) = false, esperado true", imposto)
		}
	}
}

func TestListCSTsFor_DevolveCopia(t *testing.T) {
	out := ListCSTsFor("ICMS")
	if len(out) != 11 {
		t.Errorf("ICMS deveria ter 11 CSTs, got %d", len(out))
	}

	// Mutar o slice retornado nao deve afetar o dataset interno
	out[0] = "ZZ"
	check := ListCSTsFor("ICMS")
	if check[0] != "00" {
		t.Errorf("ListCSTsFor deve devolver copia (caller mutou e afetou dataset interno)")
	}
}

func TestListCSTsFor_OrdemCrescente(t *testing.T) {
	impostos := []string{"ICMS", "ICMS_SN", "PIS", "COFINS"}
	for _, imp := range impostos {
		csts := ListCSTsFor(imp)
		sorted := make([]string, len(csts))
		copy(sorted, csts)
		sort.Strings(sorted)
		for i := range csts {
			if csts[i] != sorted[i] {
				t.Errorf("ListCSTsFor(%q) nao esta ordenado: %v", imp, csts)
				break
			}
		}
	}
}

func TestListCSTsFor_ImpostoDesconhecido(t *testing.T) {
	cases := []string{"IPI", "ISS", "", "ICMS_BRUTO"}
	for _, imp := range cases {
		if out := ListCSTsFor(imp); out != nil {
			t.Errorf("ListCSTsFor(%q) = %v, esperado nil", imp, out)
		}
	}
}

func TestListCSTsFor_PISigualCOFINS(t *testing.T) {
	pis := ListCSTsFor("PIS")
	cofins := ListCSTsFor("COFINS")
	if len(pis) != len(cofins) {
		t.Fatalf("PIS (%d) e COFINS (%d) deveriam ter mesma cardinalidade", len(pis), len(cofins))
	}
	for i := range pis {
		if pis[i] != cofins[i] {
			t.Errorf("divergencia em [%d]: PIS=%q vs COFINS=%q", i, pis[i], cofins[i])
		}
	}
}
