//go:build http

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hayride-dev/bindings/go/hayride/ai"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/repository"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server/export"
	"go.bytecodealliance.org/cm"
)

type promptReq struct {
	Message string `json:"message"`
}

type promptResp struct {
	Result []ai.Message `json:"result"`
}

func init() {
	// Initialize the context, tools, and model format
	ctx, err := ctx.New()
	if err != nil {
		log.Fatal("failed to create context:", err)
	}

	tools, err := tools.New()
	if err != nil {
		log.Fatal("failed to create tools:", err)
	}

	a, err := agents.New(
		agents.WithName("Helpful Agent"),
		agents.WithInstruction("You are a helpful assistant. Answer the user's questions to the best of your ability."),
		agents.WithContext(ctx),
		agents.WithTools(tools),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new runner instance with SSE formatting enabled
	runnerOpts := ai.RunnerOptions{
		MaxTurns: 10,
		Writer:   ai.WriterTypeSse,
	}

	runner, err := runner.New(runnerOpts)
	if err != nil {
		log.Fatal("failed to create runner:", err)
	}

	h := &handler{
		agent:  a,
		runner: runner,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", h.handlerFunc)

	// Configure the address for the spawned HTTP server
	export.ServerConfig(mux, server.ServerConfig{
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

	msg := ai.Message{
		Role: ai.RoleUser,
		Content: cm.ToList([]ai.MessageContent{
			ai.NewMessageContent(ai.Text(req.Message)),
		}),
	}

	repo := repository.New()
	path, err := repo.DownloadModel("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF/Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf")
	if err != nil {
		log.Fatal("failed to download model:", err)
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

	if _, err := h.runner.Invoke(msg, h.agent, format, graphExecutionCtxStream, w); err != nil {
		http.Error(w, "failed to invoke agent: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {}
