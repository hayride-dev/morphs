//go:build cli

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/repository"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"github.com/hayride-dev/bindings/go/wasi/cli"
	"go.bytecodealliance.org/cm"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

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
		log.Fatal("failed to create agent:", err)
	}

	// Create a new runner instance
	runner := runner.New()
	writer := cli.GetStdout(true)

	fmt.Println("What can I help with?")
	for {
		input, _ := reader.ReadString('\n')
		prompt := strings.TrimSpace(input)
		if strings.ToLower(prompt) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		msg := types.Message{
			Role: types.RoleUser,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text(input)),
			}),
		}

		err := runner.InvokeStream(msg, writer, a)
		if err != nil {
			fmt.Println("error invoking agent:", err)
			os.Exit(1)
		}

		fmt.Println("\nWhat else can I help with? (type 'exit' to quit)")
	}
}
