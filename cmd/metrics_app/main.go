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

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "# HELP sample_metric A sample metric")
		fmt.Fprintln(w, "# TYPE sample_metric gauge")
		fmt.Fprintln(w, "sample_metric{label=\"docker\"} 1.23")
		fmt.Fprintln(w, "sample_metric{label=\"https\"} 1")
	})

	server := &http.Server{
		Addr:    ":9100",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	log.Println("Metrics App listening on https://0.0.0.0:9100")
	ln, err := tls.Listen("tcp", ":9100", server.TLSConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve(ln))
}
