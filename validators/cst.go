package validators

import (
	"strings"
)

// cstsByImposto mapeia o codigo do imposto pra lista de CSTs validos
// segundo as Notas Tecnicas oficiais da SEFAZ (NT NFe 4.00 e correlatas
// pra NFCom). Listas sao ordenadas por codigo crescente.
//
// Impostos cobertos no MVP T02.08:
//   - ICMS     regime normal (Lucro Real/Presumido) — CST do anexo NT NFe
//   - ICMS_SN  Simples Nacional — CSOSN (Codigo de Situacao da Operacao SN)
//   - PIS      contribuicao pra PIS/Pasep
//   - COFINS   contribuicao pra Cofins (lista identica ao PIS por construcao)
//
// IPI e ISSQN ficam pra evolucao futura quando aparecer demanda concreta —
// nao emitimos com eles hoje (cliente piloto telecom usa NFCom + NFSe).
var cstsByImposto = map[string][]string{
	"ICMS": {
		"00", // Tributada integralmente
		"10", // Tributada e com cobranca do ICMS por substituicao tributaria
		"20", // Com reducao de base de calculo
		"30", // Isenta ou nao tributada e com cobranca do ICMS por ST
		"40", // Isenta
		"41", // Nao tributada
		"50", // Suspensao
		"51", // Diferimento
		"60", // ICMS cobrado anteriormente por ST
		"70", // Com reducao de BC e cobranca do ICMS por ST
		"90", // Outras
	},
	"ICMS_SN": {
		"101", // Tributada pelo Simples Nacional com permissao de credito
		"102", // Tributada pelo Simples Nacional sem permissao de credito
		"103", // Isencao do ICMS no SN para faixa de receita bruta
		"201", // Tributada pelo SN com permissao de credito e cobranca de ST
		"202", // Tributada pelo SN sem permissao de credito e cobranca de ST
		"203", // Isencao do ICMS no SN para faixa de receita com ST
		"300", // Imune
		"400", // Nao tributada pelo Simples Nacional
		"500", // ICMS cobrado anteriormente por ST ou por antecipacao
		"900", // Outros
	},
	// PIS e COFINS compartilham a mesma lista de CSTs por construcao da
	// legislacao (Decreto 8.426/2015 + NT NFe). Mantemos as duas chaves
	// separadas pra que o caller assine intencao no codigo.
	"PIS": {
		"01", // Operacao Tributavel com aliquota basica
		"02", // Operacao Tributavel com aliquota diferenciada
		"03", // Operacao Tributavel com aliquota por unidade de medida
		"04", // Operacao Tributavel monofasica revenda a aliquota zero
		"05", // Operacao Tributavel por ST
		"06", // Operacao Tributavel a aliquota zero
		"07", // Operacao Isenta da Contribuicao
		"08", // Operacao sem Incidencia da Contribuicao
		"09", // Operacao com Suspensao da Contribuicao
		"49", // Outras Operacoes de Saida
		"50", // Operacao com Direito a Credito vinculada exclusivamente a receita tributada no mercado interno
		"51", // Operacao com Direito a Credito vinculada exclusivamente a receita nao tributada no mercado interno
		"52", // Operacao com Direito a Credito vinculada exclusivamente a receita de exportacao
		"53", // Operacao com Direito a Credito vinculada a receitas tributadas e nao tributadas no mercado interno
		"54", // Operacao com Direito a Credito vinculada a receitas tributadas no mercado interno e de exportacao
		"55", // Operacao com Direito a Credito vinculada a receitas nao tributadas no mercado interno e de exportacao
		"56", // Operacao com Direito a Credito vinculada a receitas tributadas, nao tributadas e de exportacao
		"60", // Credito Presumido vinculado exclusivamente a receita tributada no mercado interno
		"61", // Credito Presumido vinculado exclusivamente a receita nao tributada no mercado interno
		"62", // Credito Presumido vinculado exclusivamente a receita de exportacao
		"63", // Credito Presumido vinculado a receitas tributadas e nao tributadas no mercado interno
		"64", // Credito Presumido vinculado a receitas tributadas no mercado interno e de exportacao
		"65", // Credito Presumido vinculado a receitas nao tributadas no mercado interno e de exportacao
		"66", // Credito Presumido vinculado a receitas tributadas, nao tributadas e de exportacao
		"67", // Credito Presumido outras operacoes
		"70", // Operacao de Aquisicao sem Direito a Credito
		"71", // Operacao de Aquisicao com Isencao
		"72", // Operacao de Aquisicao com Suspensao
		"73", // Operacao de Aquisicao a Aliquota Zero
		"74", // Operacao de Aquisicao sem Incidencia da Contribuicao
		"75", // Operacao de Aquisicao por ST
		"98", // Outras Operacoes de Entrada
		"99", // Outras Operacoes
	},
	"COFINS": {
		"01", "02", "03", "04", "05", "06", "07", "08", "09", "49",
		"50", "51", "52", "53", "54", "55", "56",
		"60", "61", "62", "63", "64", "65", "66", "67",
		"70", "71", "72", "73", "74", "75",
		"98", "99",
	},
}

// IsValidCST retorna true quando `cst` pertence ao dominio de CSTs validos
// do `imposto`. Comparacao da chave do imposto eh case-insensitive +
// trim; o cst eh apenas trimmed (codigos sao numericos puros).
//
// Impostos validos: "ICMS", "ICMS_SN" (Simples Nacional), "PIS", "COFINS".
// Imposto desconhecido retorna false sem panic.
//
//	validators.IsValidCST("ICMS", "00")    // true (tributada integralmente)
//	validators.IsValidCST("icms_sn", "101") // true (case-insensitive)
//	validators.IsValidCST("ICMS", "999")    // false (fora do dominio)
//	validators.IsValidCST("IPI", "00")      // false (imposto nao coberto no MVP)
func IsValidCST(imposto, cst string) bool {
	csts, ok := cstsByImposto[strings.ToUpper(strings.TrimSpace(imposto))]
	if !ok {
		return false
	}
	target := strings.TrimSpace(cst)
	for _, c := range csts {
		if c == target {
			return true
		}
	}
	return false
}

// ListCSTsFor devolve uma copia da lista de CSTs validos do imposto, em
// ordem crescente de codigo. Caller pode mutar livremente o slice
// retornado sem afetar o dataset interno.
//
// Imposto desconhecido retorna nil.
//
//	validators.ListCSTsFor("ICMS")    // ["00","10","20","30","40","41","50","51","60","70","90"]
//	validators.ListCSTsFor("IPI")     // nil
func ListCSTsFor(imposto string) []string {
	csts, ok := cstsByImposto[strings.ToUpper(strings.TrimSpace(imposto))]
	if !ok {
		return nil
	}
	out := make([]string, len(csts))
	copy(out, csts)
	return out
}
