package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server/export"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Flush headers immediately
		w.WriteHeader(http.StatusOK)

		// Send events every second for 10 seconds
		for i := 1; i <= 10; i++ {
			// Format SSE event
			data := fmt.Sprintf("data: Event %d - Hello from SSE! (timestamp: %s)\n\n", i, time.Now().Format(time.RFC3339))

			// Write the event
			w.Write([]byte(data))

			// Wait 1 second before next event (except for the last one)
			if i < 10 {
				time.Sleep(1 * time.Second)
			}
		}

		// Send a final event to indicate completion
		w.Write([]byte("data: SSE stream completed\n\n"))

	})

	export.ServerConfig(mux, server.ServerConfig{Address: "localhost:9000"})
}

func main() {}
