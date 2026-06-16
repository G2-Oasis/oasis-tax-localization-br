package reforma

import (
	"context"
	"encoding/json"
)

// Client e a abstracao do cliente da Calculadora da Reforma Tributaria.
//
// Segue o fluxo oficial de 4 passos publicado pela Receita Federal, onde
// os 3 primeiros sao chamadas HTTP ao componente (este client) e o quarto
// e a injecao do XML no NFCom (feita fora deste pacote):
//
//	1. Calcular      -> POST /api/calculadora/regime-geral
//	2. GerarXML      -> POST /api/calculadora/xml/generate?tipo=nfe
//	3. ValidarXML    -> POST /api/calculadora/xml/validate?tipo=nfe&subtipo=grupo
//	4. (fora daqui)    injeta XML retornado no XML da NFCom
//
// Implementacoes concretas:
//   - SidecarClient : HTTP real contra a calculadora rodando como sidecar.
//   - MockClient    : simula respostas para dev e testes, sem rede.
//   - DisabledClient: noop (usado quando a Reforma nao se aplica ao ambiente).
type Client interface {
	// Calcular submete os itens da operacao e recebe CBS, IBS e IS calculados.
	// A resposta preserva o JSON bruto para reuso em GerarXML.
	Calcular(ctx context.Context, req RegimeGeralRequest) (RegimeGeralResponse, error)

	// GerarXML transforma o resultado de Calcular no XML dos grupos RTC
	// (<IS>, <IBSCBS>, <ISTot>, <IBSCBSTot>) prontos para injecao no XML
	// do documento fiscal (NFCom, NF-e, CT-e). O parametro `tipo` indica o
	// tipo documental do destino (ex. "nfe").
	GerarXML(ctx context.Context, tipo string, calculo RegimeGeralResponse) ([]byte, error)

	// ValidarXML aciona a validacao estrutural do XML gerado na etapa anterior
	// antes da transmissao ao autorizador. `subtipo` distingue o nivel da
	// validacao (ex. "grupo" valida apenas os grupos RTC).
	ValidarXML(ctx context.Context, tipo, subtipo string, xml []byte) (ValidarXMLResponse, error)
}

// RegimeGeralResponse e o retorno da calculadora apos calculo dos tributos.
// O campo Raw preserva o JSON bruto porque ele e usado como entrada do
// proximo passo (GerarXML) — evita perda de informacao por serializacao
// intermediaria.
type RegimeGeralResponse struct {
	// Raw guarda o corpo JSON recebido da calculadora exatamente como veio.
	// A ser repassado para GerarXML.
	Raw json.RawMessage

	// Campos parseados serao adicionados conforme forem necessarios em
	// camadas superiores (engine, auditoria). Por enquanto mantemos apenas
	// Raw para nao criar acoplamento com uma forma especifica de resposta
	// antes de confirmar o shape via teste integrado com a calculadora.
}

// ValidarXMLResponse representa o retorno da validacao estrutural do XML.
// Quando Valido=false, Mensagem contem o corpo bruto da resposta com a lista
// de erros retornados pela calculadora.
type ValidarXMLResponse struct {
	// Valido indica se o XML passou em todas as regras estruturais.
	Valido bool

	// Mensagem traz o corpo bruto da resposta (util em caso de erro para
	// log e diagnostico).
	Mensagem string
}
