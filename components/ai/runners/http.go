//go:build http

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/hayride-dev/bindings/go/gen/types/hayride/ai/types"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/repository"
	"github.com/hayride-dev/bindings/go/wasi/net/http/handle"
)

type promptReq struct {
	msgs types.Message `json:"message"`
}

type promptResp struct {
	result types.Message `json:"result"`
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
	mux.HandleFunc("/generate", h.handlerfunc)
	handle.Handler(mux)
}

type handler struct {
	agent agents.Agent
}

func (h *handler) handlerfunc(w http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
	}

	req := &promptReq{}
	if err := json.Unmarshal(b, req); err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
	}

	response, err := h.agent.Invoke(req.msgs)
	if err != nil {
		fmt.Println("error invoking agent:", err)
		os.Exit(1)
	}

	resp := &promptResp{
		result: *response,
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
