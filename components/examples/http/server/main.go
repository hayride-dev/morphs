package main

import (
	"net/http"

	"github.com/hayride-dev/bindings/go/exports/net/http/handle"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// hayride bindings for export
	handle.Handler(mux)
}

func main() {}
