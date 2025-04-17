//go:build basic

package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/exports/ai/agent"
	"github.com/hayride-dev/bindings/go/imports/ai/ctx"
	"github.com/hayride-dev/bindings/go/imports/ai/model"
	"github.com/hayride-dev/bindings/go/shared/domain/ai"
	"github.com/hayride-dev/morphs/components/ai/tools/datetime/pkg/datetime"
)

const maxTurns = 10

var modelResource model.Model

type BasicAgent struct {
	model model.Model
	ctx   ctx.Context
}

var basicAgent BasicAgent

func init() {
	// Create model
	model, err := model.New(model.WithName("Meta-Llama-3.1-8B-Instruct-Q5_K_M.gguf"))
	if err != nil {
		panic(err)
	}

	// Create context and push system message
	instructions := `You are a tool calling agent.
	Use the tools you have to try to answer the user's question.	`

	context := ctx.NewContext()
	context.Push(&ai.Message{
		Role: ai.RoleSystem,
		Content: []ai.Content{
			&ai.TextContent{
				Text: instructions,
			},
		},
	})

	basicAgent = BasicAgent{
		model: model,
		ctx:   context,
	}

	agent.Export("basic", invoke)
}

func invoke(message []*ai.Message) ([]*ai.Message, error) {
	if err := basicAgent.ctx.Push(message...); err != nil {
		return nil, fmt.Errorf("failed to push message: %w", err)
	}

	msgs, err := basicAgent.ctx.Messages()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	// agent loop
	turns := 0
	for turns < maxTurns {
		msg, err := basicAgent.model.Compute(msgs)
		if err != nil {
			return nil, err
		}

		if err := basicAgent.ctx.Push(msg); err != nil {
			return nil, fmt.Errorf("failed to push response message: %w", err)
		}

		switch msg.Role {
		case ai.RoleAssistant:
			for _, content := range msg.Content {
				if content.Type() == "tool-input" {
					c := content.(*ai.ToolInput)
					switch c.ID {
					case "hayride:datetime@0.0.1":
						if c.Name == "date" {

							value := datetime.Date()

							basicAgent.ctx.Push(&ai.Message{
								Role: ai.RoleTool,
								Content: []ai.Content{
									&ai.ToolOutput{
										ID:          "hayride:datetime@0.0.1",
										Name:        "date",
										ContentType: "tool-output",
										Output:      value,
									},
								},
							})
						}
					default:
						return nil, fmt.Errorf("unknown tool use: %s", c.ID)
					}
				} else {
					// no tool input, end the loop
					return nil, nil
				}
			}
		}
		turns++
	}
	return nil, fmt.Errorf("max turns reached: %d", maxTurns)
}

func main() {}
