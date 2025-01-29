package proxy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	proxy "github.com/sasd13/traefik-proxy-forward"
)

func TestProxy(t *testing.T) {
	cfg := proxy.CreateConfig()
	cfg.Headers["X-Api-Key"] = "my-api-key"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := proxy.New(ctx, next, cfg, "proxy-forward-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeader(t, req, "X-Api-Key", "my-api-key")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
