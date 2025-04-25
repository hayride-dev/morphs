//go:build basic

package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/exports/ai/agents"
	"github.com/hayride-dev/bindings/go/gen/types/hayride/ai/types"
	"github.com/hayride-dev/bindings/go/imports/ai/ctx"
	"github.com/hayride-dev/bindings/go/imports/ai/models"
	"github.com/hayride-dev/morphs/components/ai/tools/datetime/pkg/datetime"
	"go.bytecodealliance.org/cm"
)

const maxTurns = 10

var modelResource models.Model

type BasicAgent struct {
	model models.Model
	ctx   ctx.Context
}

var basicAgent BasicAgent

func init() {
	// Create model
	model, err := models.New(models.WithName("Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf"))
	if err != nil {
		panic(err)
	}

	// Create context and push system message
	instructions := `You are a tool calling agent.
	Use the tools you have to try to answer the user's question.	`

	context := ctx.NewContext()
	context.Push(types.Message{
		Role: types.RoleSystem,
		Content: cm.ToList([]types.Content{
			types.ContentText(types.TextContent{
				Text: instructions,
			}),
		}),
	})

	basicAgent = BasicAgent{
		model: model,
		ctx:   context,
	}

	agents.Export(agents.WithName("basic"), agents.WithInvokeFunc(invoke))
}

func invoke(message []types.Message) ([]types.Message, error) {
	if err := basicAgent.ctx.Push(message...); err != nil {
		return nil, fmt.Errorf("failed to push message: %w", err)
	}

	msgs, err := basicAgent.ctx.Messages()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	// agent loop
	var results []types.Message
	turns := 0
	for turns < maxTurns {
		msg, err := basicAgent.model.Compute(msgs)
		if err != nil {
			return nil, err
		}

		if err := basicAgent.ctx.Push(*msg); err != nil {
			return nil, fmt.Errorf("failed to push response message: %w", err)
		}

		results = append(results, *msg)

		switch msg.Role {
		case types.RoleAssistant:
			for _, content := range msg.Content.Slice() {
				if content.String() == "tool-input" {
					c := content.ToolInput()
					switch c.ID {
					case "hayride:datetime@0.0.1":
						if c.Name == "date" {

							value := datetime.Date()

							basicAgent.ctx.Push(types.Message{
								Role: types.RoleTool,
								Content: cm.ToList([]types.Content{
									types.ContentToolOutput(types.ToolOutput{
										ID:          "hayride:datetime@0.0.1",
										Name:        "date",
										ContentType: "tool-output",
										Output:      value,
									}),
								}),
							})
						}
					default:
						return nil, fmt.Errorf("unknown tool use: %s", c.ID)
					}
				} else {
					// no tool input, end the loop
					return results, nil
				}
			}
		}
		turns++
	}
	return nil, fmt.Errorf("max turns reached: %d", maxTurns)
}

func main() {}
