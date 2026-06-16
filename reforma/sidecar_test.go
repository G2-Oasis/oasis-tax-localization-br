package reforma

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Os testes abaixo usam httptest.Server para simular a calculadora oficial,
// ou seja, testamos o comportamento HTTP do SidecarClient sem precisar do
// container Java rodando. Isso mantem a suite rapida e determinista.

func TestSidecarCalcularSucesso(t *testing.T) {
	var capturedBody []byte
	var capturedPath string
	var capturedContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedContentType = r.Header.Get("Content-Type")
		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"507f","itens":[{"numero":1,"tributos":{"cbs":{"valor":9.0}}}]}`))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)
	resp, err := client.Calcular(context.Background(), exemploRequest())
	if err != nil {
		t.Fatalf("calcular: %v", err)
	}

	if capturedPath != "/api/calculadora/regime-geral" {
		t.Errorf("path errado: %s", capturedPath)
	}
	if capturedContentType != "application/json" {
		t.Errorf("content-type errado: %s", capturedContentType)
	}

	// Confere que o corpo enviado e um JSON valido com os campos do request.
	var sent map[string]any
	if err := json.Unmarshal(capturedBody, &sent); err != nil {
		t.Fatalf("body nao eh json: %v", err)
	}
	if sent["id"] != "507f1f77bcf86cd799439011" {
		t.Errorf("id nao chegou: %v", sent["id"])
	}

	// Raw preserva o corpo bruto exatamente como veio.
	if !strings.Contains(string(resp.Raw), `"valor":9`) {
		t.Errorf("raw inesperado: %s", resp.Raw)
	}
}

func TestSidecarCalcularErroHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("erro interno"))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)
	_, err := client.Calcular(context.Background(), exemploRequest())
	if err == nil {
		t.Fatal("esperado erro HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("erro deveria mencionar 500: %v", err)
	}
}

func TestSidecarGerarXMLPassaQueryTipo(t *testing.T) {
	var capturedTipo string
	var capturedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTipo = r.URL.Query().Get("tipo")
		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rtc/>`))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)

	calc := RegimeGeralResponse{Raw: []byte(`{"id":"507f"}`)}
	xml, err := client.GerarXML(context.Background(), "nfe", calc)
	if err != nil {
		t.Fatalf("gerar xml: %v", err)
	}

	if capturedTipo != "nfe" {
		t.Errorf("tipo nao chegou na query: %s", capturedTipo)
	}
	if string(capturedBody) != `{"id":"507f"}` {
		t.Errorf("body nao repassou o Raw do calculo: %s", capturedBody)
	}
	if !strings.Contains(string(xml), "<rtc/>") {
		t.Errorf("xml inesperado: %s", xml)
	}
}

func TestSidecarValidarXMLValido(t *testing.T) {
	var capturedTipo, capturedSubtipo string
	var capturedContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTipo = r.URL.Query().Get("tipo")
		capturedSubtipo = r.URL.Query().Get("subtipo")
		capturedContentType = r.Header.Get("Content-Type")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("XML valido"))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)
	resp, err := client.ValidarXML(context.Background(), "nfe", "grupo", []byte("<rtc/>"))
	if err != nil {
		t.Fatalf("validar: %v", err)
	}
	if !resp.Valido {
		t.Fatalf("esperado valido, veio invalido: %s", resp.Mensagem)
	}
	if capturedTipo != "nfe" || capturedSubtipo != "grupo" {
		t.Errorf("query errada: tipo=%s subtipo=%s", capturedTipo, capturedSubtipo)
	}
	if capturedContentType != "application/xml" {
		t.Errorf("content-type errado: %s", capturedContentType)
	}
}

func TestSidecarValidarXMLInvalido4xx(t *testing.T) {
	// 4xx nao eh erro de transporte — eh XML invalido com mensagem no corpo.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("tag <vIBS> obrigatoria"))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)
	resp, err := client.ValidarXML(context.Background(), "nfe", "grupo", []byte("<rtc/>"))
	if err != nil {
		t.Fatalf("4xx nao deveria virar erro de transporte: %v", err)
	}
	if resp.Valido {
		t.Fatal("4xx deveria indicar invalido")
	}
	if !strings.Contains(resp.Mensagem, "vIBS") {
		t.Errorf("mensagem de validacao nao chegou: %s", resp.Mensagem)
	}
}

func TestSidecarValidarXML5xxEhErro(t *testing.T) {
	// 5xx continua sendo falha de infraestrutura, nao "xml invalido".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream offline"))
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 2*time.Second)
	_, err := client.ValidarXML(context.Background(), "nfe", "grupo", []byte("<rtc/>"))
	if err == nil {
		t.Fatal("esperado erro para 5xx")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("erro deveria mencionar 502: %v", err)
	}
}

func TestSidecarRespeitaContextDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewSidecarClient(srv.URL, 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.Calcular(ctx, exemploRequest())
	if err == nil {
		t.Fatal("esperado erro de context deadline")
	}
}

func TestSidecarUsaDefaultsQuandoVazio(t *testing.T) {
	c := NewSidecarClient("", 0)
	if c.baseURL != DefaultSidecarURL {
		t.Errorf("baseURL default errado: %s", c.baseURL)
	}
	if c.http.Timeout != DefaultSidecarTimeout {
		t.Errorf("timeout default errado: %v", c.http.Timeout)
	}
}
