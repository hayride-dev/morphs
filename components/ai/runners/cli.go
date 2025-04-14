//go:build cli

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hayride-dev/bindings/go/imports/ai/agent"
	"github.com/hayride-dev/bindings/go/imports/ai/ctx"
	"github.com/hayride-dev/bindings/go/imports/ai/model"
	"github.com/hayride-dev/bindings/go/shared/domain/ai"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What can I help with?")

	agent := agent.NewAgent()
	ctx := ctx.NewContext()
	model, err := model.New(model.WithName("Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf"))

	if err != nil {
		fmt.Println("error loading model agent:", err)
		os.Exit(1)
	}

	turn := 0
	for {
		if turn > 0 {
			fmt.Println("What else can I help with? (type 'exit' to quit)")
		}
		input, _ := reader.ReadString('\n')
		prompt := strings.TrimSpace(input)
		if strings.ToLower(prompt) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		msg := &ai.Message{
			Role: ai.RoleUser,
			Content: []ai.Content{
				&ai.TextContent{
					Text:        input,
					ContentType: "text/plain",
				},
			},
		}

		ctx.Push(msg)
		// agent should be wac'd
		// this is a problem. how to do we get the result ?
		response, err := agent.Invoke(ctx, model)

		if err != nil {
			fmt.Println("error invoking agent:", err)
			os.Exit(1)
		}

		fmt.Println("Response:", response)
		turn++
	}
}
