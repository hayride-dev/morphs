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

func init() {

	instructions := `You are a tool calling agent.
	Use the tools you have to try to answer the user's question.
	`
	agent.Export("basic", instructions, invoke)
}

func invoke(ctx ctx.Context, model model.Model) ([]*ai.Message, error) {
	msgs, err := ctx.Messages()
	if err != nil {
		return nil, err
	}
	// agent loop
	turns := 0
	for turns < maxTurns {
		msg, err := model.Compute(msgs)
		if err != nil {
			return nil, err
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

							ctx.Push(&ai.Message{
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
