package proxy

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"

	"go-insecure-proxy/pkg/config"
)

// DefaultEndpointsPath is where the list of configured routes is served
// when server.endpoints_path is not set.
const DefaultEndpointsPath = "/endpoints"

// NewProxyHandler creates a http.Handler that routes requests based on the config
// and ignores upstream SSL errors.
func NewProxyHandler(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Create a shared Transport that ignores self-signed certs
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Paths of the routes that were actually registered, in config order.
	registered := make([]string, 0, len(cfg.Routes))

	for _, route := range cfg.Routes {
		targetURL, err := url.Parse(route.Target)
		if err != nil {
			log.Printf("Error parsing target URL %q for path %q: %v", route.Target, route.Path, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Transport = insecureTransport

		// Update the director to ensure the Host header matches the target
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = targetURL.Host
			// Strict Target Routing: Use the target's path exactly, ignoring the incoming request path.
			// This prevents joining/appending.
			req.URL.Path = targetURL.Path
			req.URL.RawPath = targetURL.RawPath

			// Inject generic headers from config
			for key, value := range route.Headers {
				req.Header.Set(key, value)
			}
			log.Printf("Proxying request: %s %s -> %s", req.Method, req.RemoteAddr, req.URL.String())
		}

		log.Printf("Registering route: %s -> %s", route.Path, route.Target)
		mux.Handle(route.Path, proxy)
		registered = append(registered, route.Path)
	}

	endpointsPath := cfg.Server.EndpointsPath
	if endpointsPath == "" {
		endpointsPath = DefaultEndpointsPath
	}
	if slices.Contains(registered, endpointsPath) {
		log.Printf("Route %q is already configured; not serving the endpoints list there", endpointsPath)
	} else {
		log.Printf("Serving endpoints list on %s", endpointsPath)
		mux.Handle(endpointsPath, newEndpointsHandler(registered))
	}

	return mux
}

// newEndpointsHandler serves the configured route paths as JSON, e.g.
// {"endpoints":["/node","/opa"]}. Targets and headers are deliberately left
// out so upstream addresses and credentials are not exposed.
func newEndpointsHandler(paths []string) http.Handler {
	// Routes are fixed at startup, so the response can be built once.
	body, _ := json.Marshal(struct {
		Endpoints []string `json:"endpoints"`
	}{paths})
	body = append(body, '\n')

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}
