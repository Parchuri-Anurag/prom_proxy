package proxy

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"go-insecure-proxy/pkg/config"
)

func init() {
	log.SetOutput(io.Discard)
}

func getEndpoints(t *testing.T, h http.Handler, path string) []string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d, body %q", path, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("GET %s: Content-Type %q", path, ct)
	}
	var resp struct {
		Endpoints []string `json:"endpoints"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("GET %s: decoding %q: %v", path, rec.Body.String(), err)
	}
	return resp.Endpoints
}

func TestEndpointsListsAllRoutes(t *testing.T) {
	for _, paths := range [][]string{
		{"/node", "/opa"},
		{"/node", "/opa", "/mysql", "/redis", "/blackbox"},
	} {
		cfg := &config.Config{}
		for _, p := range paths {
			cfg.Routes = append(cfg.Routes, config.Route{Path: p, Target: "https://localhost:9100/metrics"})
		}
		got := getEndpoints(t, NewProxyHandler(cfg), DefaultEndpointsPath)
		if !slices.Equal(got, paths) {
			t.Errorf("endpoints = %v, want %v", got, paths)
		}
	}
}

func TestEndpointsNoRoutes(t *testing.T) {
	rec := httptest.NewRecorder()
	NewProxyHandler(&config.Config{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, DefaultEndpointsPath, nil))
	if got, want := rec.Body.String(), "{\"endpoints\":[]}\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestEndpointsSkipsInvalidTargets(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{
		{Path: "/node", Target: "https://localhost:9100/metrics"},
		{Path: "/broken", Target: "://bad"},
	}}
	got := getEndpoints(t, NewProxyHandler(cfg), DefaultEndpointsPath)
	if want := []string{"/node"}; !slices.Equal(got, want) {
		t.Errorf("endpoints = %v, want %v", got, want)
	}
}

func TestEndpointsCustomPath(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{EndpointsPath: "/routes"},
		Routes: []config.Route{{Path: "/node", Target: "https://localhost:9100/metrics"}},
	}
	h := NewProxyHandler(cfg)
	if got, want := getEndpoints(t, h, "/routes"), []string{"/node"}; !slices.Equal(got, want) {
		t.Errorf("endpoints = %v, want %v", got, want)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, DefaultEndpointsPath, nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET %s with custom endpoints_path: status %d, want 404", DefaultEndpointsPath, rec.Code)
	}
}

func TestEndpointsRouteTakesPrecedence(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "from upstream")
	}))
	defer upstream.Close()

	cfg := &config.Config{Routes: []config.Route{{Path: DefaultEndpointsPath, Target: upstream.URL}}}
	rec := httptest.NewRecorder()
	NewProxyHandler(cfg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, DefaultEndpointsPath, nil))
	if got := rec.Body.String(); got != "from upstream" {
		t.Errorf("body = %q, want the configured route to be proxied", got)
	}
}

func TestEndpointsRejectsNonGet(t *testing.T) {
	rec := httptest.NewRecorder()
	NewProxyHandler(&config.Config{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, DefaultEndpointsPath, nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST status = %d, want 405", rec.Code)
	}
}
