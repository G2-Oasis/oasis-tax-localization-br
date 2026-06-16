package reforma

import (
	"context"
	"errors"
)

// ErrDisabled e retornado por DisabledClient em qualquer chamada.
// Sinaliza ao chamador que a Reforma Tributaria nao esta configurada no
// ambiente atual e, portanto, nao ha calculo de CBS/IBS/IS a ser feito.
var ErrDisabled = errors.New("reformatax: client desativado")

// DisabledClient e o no-op. Usado quando o ambiente nao precisa da calculadora
// (ex. tenants fora do escopo da Reforma ou modo legado).
//
// Qualquer chamada retorna ErrDisabled — cabe ao FiscalEngine decidir como
// reagir (seguir sem RTC, logar warning, etc). Assim a ausencia da calculadora
// e explicita, em vez de virar uma resposta silenciosamente vazia.
type DisabledClient struct{}

func NewDisabledClient() *DisabledClient { return &DisabledClient{} }

func (DisabledClient) Calcular(ctx context.Context, req RegimeGeralRequest) (RegimeGeralResponse, error) {
	return RegimeGeralResponse{}, ErrDisabled
}

func (DisabledClient) GerarXML(ctx context.Context, tipo string, calculo RegimeGeralResponse) ([]byte, error) {
	return nil, ErrDisabled
}

func (DisabledClient) ValidarXML(ctx context.Context, tipo, subtipo string, xml []byte) (ValidarXMLResponse, error) {
	return ValidarXMLResponse{}, ErrDisabled
}

// Garantia estatica de que DisabledClient implementa Client.
var _ Client = (*DisabledClient)(nil)
