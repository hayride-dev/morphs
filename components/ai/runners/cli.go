//go:build cli

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hayride-dev/bindings/go/gen/types/hayride/ai/types"
	"github.com/hayride-dev/bindings/go/imports/ai/agents"
	"go.bytecodealliance.org/cm"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What can I help with?")

	a := agents.NewAgent()

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

		msg := types.Message{
			Role: types.RoleUser,
			Content: cm.ToList([]types.Content{
				types.ContentText(types.TextContent{
					Text:        input,
					ContentType: "text/plain",
				}),
			}),
		}

		response, err := a.Invoke([]types.Message{msg})

		if err != nil {
			fmt.Println("error invoking agent:", err)
			os.Exit(1)
		}

		fmt.Println("Response:", response)
		turn++
	}
}
