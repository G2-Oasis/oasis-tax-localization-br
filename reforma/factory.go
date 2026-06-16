package reforma

import (
	"fmt"
	"strings"
	"time"
)

// Mode define qual implementacao da Calculadora usar.
type Mode string

const (
	ModeDisabled Mode = ""        // sem calculadora (ambiente legado ou fora do escopo da Reforma)
	ModeMock     Mode = "mock"    // MockClient em memoria (dev/teste)
	ModeSidecar  Mode = "sidecar" // SidecarClient HTTP (producao local)
)

// New seleciona a implementacao do Client conforme mode.
// Valores invalidos caem em ModeDisabled (fail-safe: ausencia e melhor que
// resultado silenciosamente errado).
func New(mode, baseURL string, timeout time.Duration) (Client, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(mode))) {
	case ModeDisabled:
		return NewDisabledClient(), nil
	case ModeMock:
		return NewMockClient(), nil
	case ModeSidecar:
		return NewSidecarClient(baseURL, timeout), nil
	default:
		return nil, fmt.Errorf("reformatax: modo invalido %q (esperado: mock, sidecar ou vazio)", mode)
	}
}
