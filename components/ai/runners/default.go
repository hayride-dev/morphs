package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
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

func (r *defaultRunner) Invoke(message types.Message, agent agents.Agent, format models.Format, model graph.GraphExecutionContextStream, w io.Writer) ([]types.Message, error) {
	messages := make([]types.Message, 0)

	if err := agent.Push(message); err != nil {
		return nil, fmt.Errorf("failed to push message to agent: %w", err)
	}
	toolCall := false
	for i := 0; i <= maxturn; i++ {
		history, err := agent.Context()
		if err != nil {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}
		// Format encode the messages
		data, err := format.Encode(history...)
		if err != nil {
			return nil, fmt.Errorf("failed to encode context messages: %w", err)
		}

		// Call Graph Compute
		d := graph.TensorDimensions(cm.ToList([]uint32{1}))
		td := graph.TensorData(cm.ToList(data))
		t := graph.NewTensor(d, graph.TensorTypeU8, td)
		inputs := []graph.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}

		namedTensorStream, err := model.Compute(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to compute graph: %w", err)
		}

		// read the output from the stream
		stream := namedTensorStream.F1
		ts := graph.TensorStream(stream)
		//
		text := make([]byte, 0)
		part := make([]byte, 256)
		for {
			len, err := ts.Read(part)
			if len == 0 || err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("failed to read from tensor stream: %w", err)
			}
			text = append(text, part[:len]...)
			// Decode Message
			msg, err := format.Decode(text) // Gets us our API Message
			if err != nil {
				if _, ok := err.(*models.PartialDecodeError); ok {
					// continue reading
					continue
				}
				return nil, err
			}
			agent.Push(*msg)
			// Add the message to the messages list
			messages = append(messages, *msg)
			// write the message bytes as json to the writer if provided ( i.e stream )
			if w != nil {
				b, err := json.Marshal(msg)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal message: %w", err)
				}
				w.Write(b)
			}

			// check for tool call
			if msg.Role == types.RoleAssistant {
				for _, c := range msg.Content.Slice() {
					if c.String() == "tool-input" {
						toolResult, err := agent.Execute(*c.ToolInput())
						if err != nil {
							return nil, fmt.Errorf("failed to call tool: %w", err)
						}
						toolCallMessage := types.Message{Role: types.RoleTool, Content: cm.ToList([]types.MessageContent{types.NewMessageContent(*toolResult)})}
						agent.Push(toolCallMessage)
						messages = append(messages, toolCallMessage)
						toolCall = true
					}
				}
			}
		}
		if toolCall {
			continue // If we had a tool call, we need to continue the loop to process the tool result
		}
		messages[len(messages)-1].Final = true // Mark the last message as final
		break
	}
	return messages, nil
}

func main() {}
