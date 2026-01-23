package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"

	"go-insecure-proxy/pkg/certs"
	"go-insecure-proxy/pkg/config"
	"go-insecure-proxy/pkg/proxy"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		println("Insecure Proxy v0.1.0")
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handler := proxy.NewProxyHandler(cfg)

	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: handler,
	}

	if cfg.Server.TLS {
		log.Printf("Starting HTTPS server on %s", cfg.Server.Addr)

		var tlsConfig *tls.Config

		// If cert files are provided, load them. Otherwise, generate self-signed.
		if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.Server.CertFile, cfg.Server.KeyFile)
			if err != nil {
				log.Fatalf("Failed to load key pair: %v", err)
			}
			tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		} else {
			log.Println("No certificate files provided, generating self-signed certificate...")
			cert, err := certs.GenerateSelfSignedCert()
			if err != nil {
				log.Fatalf("Failed to generate self-signed certificate: %v", err)
			}
			tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		}

		server.TLSConfig = tlsConfig
		// We use server.ServeTLS instead of ListenAndServeTLS because we manually configured the TLSConfig
		// with an in-memory or loaded certificate.
		// However, ServeTLS expects a listener, so we create one.

		ln, err := tls.Listen("tcp", server.Addr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", server.Addr, err)
		}

		if err := server.Serve(ln); err != nil {
			log.Fatalf("Server failed: %v", err)
		}

	} else {
		log.Printf("Starting HTTP server on %s", cfg.Server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}
