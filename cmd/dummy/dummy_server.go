package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"go-insecure-proxy/pkg/certs"
)

func main() {
	cert, err := certs.GenerateSelfSignedCert()
	if err != nil {
		log.Fatal(err)
	}

	// Listen on specific IP to test interface binding
	addr := "10.136.21.11:9443"

	server := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello from Dummy Server on Specific IP %s!", addr)
		}),
	}

	log.Printf("Dummy server listening on %s", addr)
	ln, err := tls.Listen("tcp", addr, server.TLSConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve(ln))
}
