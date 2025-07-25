//go:build http

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/repository"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server/export"
	"go.bytecodealliance.org/cm"
)

type promptReq struct {
	Message string `json:"message"`
}

type promptResp struct {
	Result []types.Message `json:"result"`
}

func init() {

	repo := repository.New()
	path, err := repo.DownloadModel("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF/Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf")
	if err != nil {
		log.Fatal("failed to download model:", err)
	}

	// Initialize the context, tools, and model format
	ctx, err := ctx.New()
	if err != nil {
		log.Fatal("failed to create context:", err)
	}

	tools, err := tools.New()
	if err != nil {
		log.Fatal("failed to create tools:", err)
	}

	format, err := models.New()
	if err != nil {
		log.Fatal("failed to create model format:", err)
	}

	// host provides a graph stream
	inferenceStream, err := graph.LoadByName(path)
	if err != nil {
		log.Fatal("failed to load graph:", err)
	}

	graphExecutionCtxStream, err := inferenceStream.InitExecutionContextStream()
	if err != nil {
		log.Fatal("failed to initialize graph execution context stream:", err)
	}

	a, err := agents.New(
		format, graphExecutionCtxStream,
		agents.WithName("Helpful Agent"),
		agents.WithInstruction("You are a helpful assistant. Answer the user's questions to the best of your ability."),
		agents.WithContext(ctx),
		agents.WithTools(tools),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new runner instance
	runner := runner.New()

	h := &handler{
		agent:  a,
		runner: runner,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", h.handlerFunc)

	// Configure the address for the spawned HTTP server
	export.ServerConfig(mux, types.ServerConfig{
		Address: "http://localhost:8083",
	})
}

type handler struct {
	agent  agents.Agent
	runner runner.Runner
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
		Content: cm.ToList([]types.MessageContent{
			types.NewMessageContent(types.Text(req.Message)),
		}),
	}

	response, err := h.runner.Invoke(msg, h.agent)
	if err != nil {
		http.Error(w, "failed to invoke agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &promptResp{
		Result: response,
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
