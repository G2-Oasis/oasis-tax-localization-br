package reforma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultSidecarURL endereco padrao onde o container da calculadora roda
// ao lado do OTAX-001 (host interno do Docker Compose ou localhost em dev).
const DefaultSidecarURL = "http://localhost:8080"

// DefaultSidecarTimeout timeout generoso para a chamada inteira. O chamador
// pode apertar isso via ctx com deadline proprio quando for o caso.
const DefaultSidecarTimeout = 30 * time.Second

// SidecarClient conversa com a calculadora oficial da Receita rodando como
// processo sidecar (container Docker ou JAR) no mesmo host.
//
// Importante: todas as chamadas sao POST com corpo/query conforme a API
// oficial. Erros de rede e HTTP >=500 viram `error`; HTTP 4xx em ValidarXML
// eh tratado como "XML invalido" (e nao como erro de transporte).
type SidecarClient struct {
	baseURL string
	http    *http.Client
}

// NewSidecarClient cria um client apontado para baseURL.
// Se baseURL for vazio, usa DefaultSidecarURL. Se timeout for zero, usa
// DefaultSidecarTimeout.
func NewSidecarClient(baseURL string, timeout time.Duration) *SidecarClient {
	if baseURL == "" {
		baseURL = DefaultSidecarURL
	}
	if timeout <= 0 {
		timeout = DefaultSidecarTimeout
	}
	return &SidecarClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

// Calcular chama POST /api/calculadora/regime-geral.
func (c *SidecarClient) Calcular(ctx context.Context, req RegimeGeralRequest) (RegimeGeralResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return RegimeGeralResponse{}, fmt.Errorf("reformatax: marshal request: %w", err)
	}

	endpoint := c.baseURL + "/api/calculadora/regime-geral"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return RegimeGeralResponse{}, fmt.Errorf("reformatax: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return RegimeGeralResponse{}, fmt.Errorf("reformatax: calculadora unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return RegimeGeralResponse{}, fmt.Errorf("reformatax: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return RegimeGeralResponse{}, fmt.Errorf("reformatax: calcular retornou %d: %s", resp.StatusCode, truncate(respBody, 500))
	}

	return RegimeGeralResponse{Raw: respBody}, nil
}

// GerarXML chama POST /api/calculadora/xml/generate?tipo={tipo}.
// Passa o Raw do calculo anterior como corpo (JSON) e espera XML na resposta.
func (c *SidecarClient) GerarXML(ctx context.Context, tipo string, calculo RegimeGeralResponse) ([]byte, error) {
	u, err := url.Parse(c.baseURL + "/api/calculadora/xml/generate")
	if err != nil {
		return nil, fmt.Errorf("reformatax: parse url: %w", err)
	}
	q := u.Query()
	q.Set("tipo", tipo)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(calculo.Raw))
	if err != nil {
		return nil, fmt.Errorf("reformatax: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/xml")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("reformatax: calculadora unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reformatax: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reformatax: gerar xml retornou %d: %s", resp.StatusCode, truncate(respBody, 500))
	}

	return respBody, nil
}

// ValidarXML chama POST /api/calculadora/xml/validate?tipo={tipo}&subtipo={subtipo}.
// 2xx -> XML valido. 4xx -> XML invalido (mensagem no corpo). 5xx/rede -> erro.
func (c *SidecarClient) ValidarXML(ctx context.Context, tipo, subtipo string, xml []byte) (ValidarXMLResponse, error) {
	u, err := url.Parse(c.baseURL + "/api/calculadora/xml/validate")
	if err != nil {
		return ValidarXMLResponse{}, fmt.Errorf("reformatax: parse url: %w", err)
	}
	q := u.Query()
	q.Set("tipo", tipo)
	if subtipo != "" {
		q.Set("subtipo", subtipo)
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(xml))
	if err != nil {
		return ValidarXMLResponse{}, fmt.Errorf("reformatax: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/xml")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ValidarXMLResponse{}, fmt.Errorf("reformatax: calculadora unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ValidarXMLResponse{}, fmt.Errorf("reformatax: read response: %w", err)
	}

	// 5xx continua sendo falha de infraestrutura; 2xx valido; 4xx invalido.
	if resp.StatusCode >= 500 {
		return ValidarXMLResponse{}, fmt.Errorf("reformatax: validar xml retornou %d: %s", resp.StatusCode, truncate(respBody, 500))
	}

	return ValidarXMLResponse{
		Valido:   resp.StatusCode >= 200 && resp.StatusCode < 300,
		Mensagem: string(respBody),
	}, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

// Garantia estatica de que SidecarClient implementa Client.
var _ Client = (*SidecarClient)(nil)
