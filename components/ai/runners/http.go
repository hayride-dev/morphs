//go:build http

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hayride-dev/bindings/go/exports/net/http/handle"
	"github.com/hayride-dev/bindings/go/imports/ai/agent"
	"github.com/hayride-dev/bindings/go/imports/ai/ctx"
	"github.com/hayride-dev/bindings/go/imports/ai/model"
	"github.com/hayride-dev/bindings/go/shared/domain/ai"
)

type promptReq struct {
	msgs []*ai.Message `json:"messages"`
}

type promptResp struct {
	result []*ai.Message `json:"result"`
}

func init() {
	model, err := model.New(model.WithName("Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf"))
	if err != nil {
		fmt.Println("error loading model agent:", err)
		os.Exit(1)
	}
	h := &handler{
		agent: agent.NewAgent(),
		model: model,
		ctx:   ctx.NewContext(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", h.handlerfunc)
	handle.Handler(mux)
}

type handler struct {
	agent agent.Agent
	model model.Model
	ctx   ctx.Context
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
	h.ctx.Push(req.msgs...)

	response, err := h.agent.Invoke(h.ctx, h.model)
	if err != nil {
		fmt.Println("error invoking agent:", err)
		os.Exit(1)
	}

	resp := &promptResp{
		result: response,
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
