package main

import (
	"net/http"

	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server/export"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	export.ServerConfig(mux, server.ServerConfig{Address: "localhost:9000"})
}

func main() {}
