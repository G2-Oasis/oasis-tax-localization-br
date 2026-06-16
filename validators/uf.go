package validators

import "strings"

// ufIBGECodes mapeia a sigla da Unidade Federativa pro codigo IBGE de 2
// digitos. Tabela canonica da localizacao fiscal BR — todo lugar que
// precisa do cUF (chave de acesso 44d, <consStatServ> da NFe,
// <enderEmit>/<enderDest>, <cMunFG>, etc) deve consultar esta estrutura.
//
// Antes do T02.07 a tabela vivia duplicada em:
//   - OTAX-001 internal/service/numbering.go (ufIBGECode)
//   - OTAX-001 internal/service/nfcom_envelope.go (mesma tabela)
//   - OTAX-002 internal/domain/uf.go (ufIBGECodes, ja consolidada em BL-145)
//
// Centralizar aqui na biblioteca de localizacao garante single source of
// truth — se o IBGE adicionar nova UF (improvavel mas possivel pra
// territorios), atualizar um lugar so.
var ufIBGECodes = map[string]string{
	"AC": "12", "AL": "27", "AP": "16", "AM": "13", "BA": "29",
	"CE": "23", "DF": "53", "ES": "32", "GO": "52", "MA": "21",
	"MT": "51", "MS": "50", "MG": "31", "PA": "15", "PB": "25",
	"PR": "41", "PE": "26", "PI": "22", "RJ": "33", "RN": "24",
	"RS": "43", "RO": "11", "RR": "14", "SC": "42", "SP": "35",
	"SE": "28", "TO": "17",
}

// ibgeToUF eh o inverso de ufIBGECodes — construido na inicializacao do
// pacote pra que IBGECodeToUF seja O(1) sem manter as duas tabelas em
// sincronia manual.
var ibgeToUF = func() map[string]string {
	out := make(map[string]string, len(ufIBGECodes))
	for sigla, cUF := range ufIBGECodes {
		out[cUF] = sigla
	}
	return out
}()

// UFToIBGECode traduz uma sigla de UF (case-insensitive, com trim) pro
// codigo IBGE de 2 digitos. Retorna ("", false) quando a sigla nao e uma
// das 27 UFs conhecidas.
//
//	cUF, ok := validators.UFToIBGECode("ES")  // "32", true
//	cUF, ok := validators.UFToIBGECode(" rs") // "43", true (case-insensitive, trim)
//	cUF, ok := validators.UFToIBGECode("XX")  // "", false
func UFToIBGECode(uf string) (string, bool) {
	code, ok := ufIBGECodes[strings.ToUpper(strings.TrimSpace(uf))]
	return code, ok
}

// IBGECodeToUF traduz um codigo IBGE de 2 digitos pra sigla da UF. Util
// pra interpretar respostas da SEFAZ que devolvem <cUF> em vez de sigla
// (ex: retornos de NFeStatusServico4 ou retConsSitNFe). Aceita entrada
// com trim — o codigo IBGE eh ja numerico, case nao aplica.
//
//	sigla, ok := validators.IBGECodeToUF("32") // "ES", true
//	sigla, ok := validators.IBGECodeToUF("99") // "", false
func IBGECodeToUF(cUF string) (string, bool) {
	sigla, ok := ibgeToUF[strings.TrimSpace(cUF)]
	return sigla, ok
}
