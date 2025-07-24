package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner/export"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

const maxturn = 10

var _ runner.Runner = (*defaultRunner)(nil)

type defaultRunner struct{}

func init() {
	runner := &defaultRunner{}
	export.Runner(runner)
}

func (r *defaultRunner) Invoke(message types.Message, agent agents.Agent) ([]types.Message, error) {

	var messages []types.Message
	currentMessage := message
	for i := 0; i <= maxturn; i++ {
		msg, err := agent.Compute(currentMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to compute message: %w", err)
		}

		// Add the message to the messages list
		messages = append(messages, *msg)

		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolResult, err := agent.Execute(*c.ToolInput())
					if err != nil {
						return nil, fmt.Errorf("failed to call tool: %w", err)
					}
					calledTool = true

					toolCall := types.Message{Role: types.RoleTool, Content: cm.ToList([]types.MessageContent{types.NewMessageContent(*toolResult)})}

					// Add the tool call to the messages
					messages = append(messages, toolCall)

					// re-compute with the tool output
					currentMessage = toolCall
				default:
					// If the content is not a tool input, we can just continue
					continue
				}
			}
		default:
			// the role should always be an assistant
			return nil, fmt.Errorf("unexpected role: %s", msg.Role)
		}
		if !calledTool {
			// assuming if the agent is not requesting a tool call, it is the final message
			break
		}
	}
	return messages, nil
}

func (r *defaultRunner) InvokeStream(message types.Message, writer io.Writer, agent agents.Agent) error {

	currentMessage := message
	for i := 0; i <= maxturn; i++ {
		msg, err := agent.Compute(currentMessage)
		if err != nil {
			return fmt.Errorf("failed to compute message: %w", err)
		}

		// Check for a tool call
		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolResult, err := agent.Execute(*c.ToolInput())
					if err != nil {
						return fmt.Errorf("failed to call tool: %w", err)
					}
					calledTool = true
					// Push the tool output to the context and re-compute with the tool output
					currentMessage = types.Message{
						Role:    types.RoleTool,
						Content: cm.ToList([]types.MessageContent{types.NewMessageContent(*toolResult)}),
					}
				default:
					// If the content is not a tool input, we can just continue
					continue
				}
			}
		default:
			// the role should always be an assistant
			return fmt.Errorf("unexpected role: %s", msg.Role)
		}

		if !calledTool {
			// Write full message to the output stream
			bytes, err := json.Marshal(msg)
			if err != nil {
				return fmt.Errorf("failed to marshal final message: %w", err)
			}
			_, err = writer.Write(bytes)
			if err != nil {
				return fmt.Errorf("failed to write final message: %w", err)
			}

			// assuming if the agent is not requesting a tool call, it is the final message
			break
		}
	}

	return nil
}

func main() {}
