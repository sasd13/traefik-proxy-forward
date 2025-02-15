// Package traefik_proxy_forward_test provides the tests.
// revive:disable-next-line var-naming.
package traefik_proxy_forward_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	proxy "github.com/sasd13/traefik-proxy-forward"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	cfg := proxy.CreateConfig()
	cfg.Headers["X-Api-Key"] = "DEMO_KEY"

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	handler, err := proxy.New(ctx, next, cfg, "proxy-forward-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	req.Header.Set("Location", "https://api.nasa.gov/planetary/apod")
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, 200, recorder.Result().StatusCode)
}
