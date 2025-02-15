package traefik_proxy_forward

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: make(map[string]string),
	}
}

// ProxyForward a Demo plugin.
type ProxyForward struct {
	headers  map[string][]string
	next     http.Handler
	name     string
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	headers := make(map[string][]string)

	for key, value := range config.Headers {
		headers[key] = []string{value}
	}

	return &ProxyForward{
		headers:  headers,
		next:     next,
		name:     name,
	}, nil
}

func (p *ProxyForward) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	location := r.Header.Get("Location")
	if location == "" {
		p.next.ServeHTTP(rw, r)
		return
	}

	log.Printf("Forwarding request to: %s", location)

	// Read and copy the request body
	reqBody, err := p.readRequestBody(rw, r)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(rw, "Failed to forward request", http.StatusInternalServerError)
		return
	}

	// Create a new request to the Location header
	req, err := http.NewRequestWithContext(r.Context(), r.Method, location, bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		http.Error(rw, "Failed to forward request", http.StatusInternalServerError)
		return
	}

	// Copy original headers to the new request
	p.copyHeadersToRequest(r.Header, req)

	// Copy config headers to the new request
	p.copyHeadersToRequest(p.headers, req)

	// Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		http.Error(rw, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers and status code
	p.copyHeadersToResponse(resp.Header, &rw)
	rw.WriteHeader(resp.StatusCode)

	// Copy the response body
	if _, err := io.Copy(rw, resp.Body); err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}
}

func (p *ProxyForward) readRequestBody(rw http.ResponseWriter, r *http.Request) ([]byte, error) {
	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		// Reset request body for further use
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return body, nil
}

func (p *ProxyForward) copyHeadersToRequest(header http.Header, r *http.Request) {
	for key, values := range header {
		for _, value := range values {
			if value == "" {
				r.Header.Del(key)
			} else {
				r.Header.Set(key, value)
			}
		}
	}
}

func (p *ProxyForward) copyHeadersToResponse(header http.Header, rw *http.ResponseWriter) {
	for key, values := range header {
		for _, value := range values {
			(*rw).Header().Set(key, value)
		}
	}
}
