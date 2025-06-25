//go:build http

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hayride-dev/bindings/go/gen/types/hayride/ai/types"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/repository"
	"github.com/hayride-dev/bindings/go/hayride/net/http/server"
	"go.bytecodealliance.org/cm"
)

type promptReq struct {
	Message string `json:"message"`
}

type promptResp struct {
	Result types.Message `json:"result"`
}

func init() {
	path, err := repository.Download("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF/Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf")
	if err != nil {
		log.Fatal("failed to download model:", err)
	}

	a, err := agents.New(
		agents.WithModel(path),
		agents.WithName("Helpful Agent"),
		agents.WithInstruction("You are a helpful assistant. Answer the user's questions to the best of your ability."),
	)
	if err != nil {
		log.Fatal(err)
	}
	h := &handler{
		agent: a,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", h.handlerFunc)

	// Configure the address for the spawned HTTP server
	server.Export(mux, server.Config{
		Address: "http://localhost:8083",
	})
}

type handler struct {
	agent agents.Agent
}

func (h *handler) handlerFunc(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req promptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received message:", req.Message)

	msg := types.Message{
		Role: types.RoleUser,
		Content: cm.ToList([]types.Content{
			types.ContentText(types.TextContent{
				Text:        req.Message,
				ContentType: "text/plain",
			}),
		}),
	}

	response, err := h.agent.Invoke(msg)
	if err != nil {
		http.Error(w, "failed to invoke agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &promptResp{
		Result: *response,
	}

	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal response: "+err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func main() {}
