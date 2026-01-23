package proxy

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go-insecure-proxy/pkg/config"
)

// NewProxyHandler creates a http.Handler that routes requests based on the config
// and ignores upstream SSL errors.
func NewProxyHandler(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Create a shared Transport that ignores self-signed certs
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

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

			// Inject generic headers from config
			for key, value := range route.Headers {
				req.Header.Set(key, value)
			}
		}

		log.Printf("Registering route: %s -> %s", route.Path, route.Target)
		mux.Handle(route.Path, proxy)
	}

	return mux
}
