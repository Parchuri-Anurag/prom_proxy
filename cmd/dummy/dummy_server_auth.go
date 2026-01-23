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

	server := &http.Server{
		Addr: ":9443",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Unauthorized: No Authorization header", http.StatusUnauthorized)
				return
			}
			fmt.Fprintf(w, "Authorized! Header value: %s", auth)
		}),
	}

	log.Println("Auth Dummy server listening on :9443")
	ln, err := tls.Listen("tcp", ":9443", server.TLSConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve(ln))
}
