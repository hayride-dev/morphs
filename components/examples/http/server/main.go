package main

import (
	"net/http"

	"github.com/hayride-dev/bindings/go/exports/net/http/server"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	server.Export(mux, server.Config{Address: "localhost:9000"})
}

func main() {}
