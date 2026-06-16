// Package reformatax integra o OTAX-001 com a Calculadora oficial da Reforma
// Tributaria sobre o Consumo, publicada pela Receita Federal.
//
// A calculadora oficial e uma aplicacao Spring Boot Java que roda como sidecar
// local (Docker ou JAR). Este pacote encapsula as chamadas HTTP ao componente
// rodando em http://localhost:8080 por padrao.
//
// Fonte oficial: https://piloto-cbs.tributos.gov.br/servico/calculadora-consumo/
package reforma

// RegimeGeralRequest e o corpo da requisicao para POST /api/calculadora/regime-geral.
// Representa um documento fiscal (a ser) emitido com os itens a serem tributados.
type RegimeGeralRequest struct {
	// ID identifica a operacao de forma unica (recomendado: UUID ou ObjectID).
	// Usado para rastreabilidade do calculo no log da calculadora.
	ID string `json:"id"`

	// Versao do contrato/JSON aceito pela calculadora. Atualmente "1.0.0".
	Versao string `json:"versao"`

	// DataHoraEmissao em formato ISO 8601 com timezone (ex. "2027-01-01T03:00:00-03:00").
	// Define qual aliquota-teste/normal a calculadora aplicara (transicao 2026-2033).
	DataHoraEmissao string `json:"dataHoraEmissao"`

	// Municipio e o codigo IBGE (7 digitos) do municipio de ocorrencia do fato
	// gerador. Relevante para IBS (parcela municipal).
	Municipio int `json:"municipio"`

	// UF e a sigla da UF de emissao (ex. "RS"). Relevante para IBS (parcela estadual).
	UF string `json:"uf"`

	// Itens da operacao. A calculadora retorna CBS/IBS/IS por item.
	Itens []ItemInput `json:"itens"`
}

// ItemInput representa um item da operacao a ser tributado.
//
// Ha dois grupos opcionais aninhados que controlam o comportamento do calculo:
//   - TributacaoRegular: regras especificas para CBS/IBS (aliquota reduzida,
//     isencao, imunidade, etc.) referenciadas por CST + cClassTrib.
//   - ImpostoSeletivo: aplica somente quando o item esta sujeito ao IS
//     (ex. bebidas alcoolicas, fumo, etc. — chamado popularmente de "imposto do pecado").
type ItemInput struct {
	// Numero sequencial do item no documento (1-based).
	Numero int `json:"numero"`

	// NCM (Nomenclatura Comum do Mercosul) do item, 8 digitos.
	NCM string `json:"ncm"`

	// Quantidade comercializada.
	Quantidade float64 `json:"quantidade"`

	// Unidade de medida (ex. "VN" = valor, "UN" = unidade, "KG", "LT", etc.).
	Unidade string `json:"unidade"`

	// CST = Codigo de Situacao Tributaria da Reforma (diferente do CST do ICMS).
	// Identifica o regime aplicavel (tributado, isento, imune, suspenso, etc.).
	CST string `json:"cst"`

	// BaseCalculo do item em reais (valor sobre o qual as aliquotas sao aplicadas).
	BaseCalculo float64 `json:"baseCalculo"`

	// CClassTrib = Codigo de Classificacao Tributaria. Chave para lookup de
	// regras especificas na base da calculadora (combinacao NCM + CST + contexto).
	CClassTrib string `json:"cClassTrib"`

	// TributacaoRegular opcional — usado quando ha regime especial de CBS/IBS
	// (ex. aliquota reduzida para medicamentos, isencao para cesta basica).
	TributacaoRegular *TributacaoRegular `json:"tributacaoRegular,omitempty"`

	// ImpostoSeletivo opcional — usado quando o item esta sujeito ao IS.
	ImpostoSeletivo *ImpostoSeletivo `json:"impostoSeletivo,omitempty"`
}

// TributacaoRegular define o regime especial de CBS/IBS aplicavel ao item.
// Presente apenas em itens com tratamento diferenciado (reducoes, isencoes).
type TributacaoRegular struct {
	// CST do regime especifico (ex. "200" = aliquota reduzida).
	CST string `json:"cst"`

	// CClassTrib do regime especifico.
	CClassTrib string `json:"cClassTrib"`
}

// ImpostoSeletivo informa dados para calculo do IS (tributo sobre produtos
// nocivos a saude ou ao meio ambiente). So preencher quando aplicavel.
type ImpostoSeletivo struct {
	CST         string  `json:"cst"`
	BaseCalculo float64 `json:"baseCalculo"`
	CClassTrib  string  `json:"cClassTrib"`
	Unidade     string  `json:"unidade"`
	Quantidade  float64 `json:"quantidade"`

	// ImpostoInformado permite o emitente declarar um valor de IS ja calculado
	// externamente (raro — normalmente zero pra calculadora computar).
	ImpostoInformado float64 `json:"impostoInformado"`
}
