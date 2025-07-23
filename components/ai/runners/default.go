package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
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
	ctx := agent.Context()
	format := agent.Format()
	tools := agent.Tools()
	g := agent.Graph()

	err := ctx.Push(message)
	if err != nil {
		return nil, fmt.Errorf("failed to push message to context: %w", err)
	}

	var messages []types.Message
	for i := 0; i <= maxturn; i++ {
		msgs, err := ctx.Messages()
		if err != nil {
			return nil, fmt.Errorf("failed to get messages from context: %w", err)
		}

		data, err := format.Encode(msgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to encode messages: %w", err)
		}

		d := graph.TensorDimensions(cm.ToList([]uint32{1}))
		td := graph.TensorData(cm.ToList(data))
		t := graph.NewTensor(d, graph.TensorTypeU8, td)
		inputs := []graph.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}
		namedTensorStream, err := g.Compute(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to compute graph: %w", err)
		}

		stream := namedTensorStream.F1
		ts := graph.TensorStream(stream)
		// read the output from the stream
		text := make([]byte, 0)
		for {
			// Read up to 100 bytes from the output
			// to get any tokens that have been generated and push to socket
			p := make([]byte, 100)
			len, err := ts.Read(p)
			if len == 0 || err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("failed to read from tensor stream: %w", err)
			}
			text = append(text, p[:len]...)
		}

		msg, err := format.Decode(text)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if err := ctx.Push(*msg); err != nil {
			return nil, fmt.Errorf("failed to push message to context: %w", err)
		}

		// Add the message to the messages list
		messages = append(messages, *msg)

		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolResult, err := tools.Call(*c.ToolInput())
					if err != nil {
						return nil, fmt.Errorf("failed to call tool: %w", err)
					}
					calledTool = true

					toolCall := types.Message{Role: types.RoleTool, Content: cm.ToList([]types.MessageContent{types.NewMessageContent(*toolResult)})}

					// Add the tool call to the messages
					messages = append(messages, toolCall)

					// Push the tool output to the context and re-compute with the tool output
					ctx.Push(toolCall)
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
	ctx := agent.Context()
	format := agent.Format()
	tools := agent.Tools()
	g := agent.Graph()

	if err := ctx.Push(message); err != nil {
		return fmt.Errorf("failed to push message to context: %w", err)
	}

	finalMsg := &types.Message{Role: types.RoleAssistant, Content: cm.ToList([]types.MessageContent{types.NewMessageContent(types.Text("agent yielded no response"))})}

	for i := 0; i <= maxturn; i++ {
		msgs, err := ctx.Messages()
		if err != nil {
			return fmt.Errorf("failed to get messages from context: %w", err)
		}

		data, err := format.Encode(msgs...)
		if err != nil {
			return fmt.Errorf("failed to encode messages: %w", err)
		}

		d := graph.TensorDimensions(cm.ToList([]uint32{1}))
		td := graph.TensorData(cm.ToList(data))
		t := graph.NewTensor(d, graph.TensorTypeU8, td)
		inputs := []graph.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}
		namedTensorStream, err := g.Compute(inputs)
		if err != nil {
			return fmt.Errorf("failed to compute graph: %w", err)
		}

		stream := namedTensorStream.F1
		ts := graph.TensorStream(stream)
		// read the output from the stream
		text := make([]byte, 0)
		for {
			// Read up to 100 bytes from the output
			// to get any tokens that have been generated and push to socket
			p := make([]byte, 100)
			len, err := ts.Read(p)
			if len == 0 || err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("failed to read from tensor stream: %w", err)
			}
			text = append(text, p[:len]...)

			// TODO:: Optionally write RAW output to the writer
			// this would result in data getting back to the client faster
			// additionally once the full message is read in, we will decode it
			// and write the full typed message.
			// For this to work cleanly, we need a new message content type, potentially role type as well.
		}

		msg, err := format.Decode(text)
		if err != nil {
			return fmt.Errorf("failed to decode message: %w", err)
		}

		if err := ctx.Push(*msg); err != nil {
			return fmt.Errorf("failed to push message to context: %w", err)
		}

		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolResult, err := tools.Call(*c.ToolInput())
					if err != nil {
						return fmt.Errorf("failed to call tool: %w", err)
					}
					calledTool = true
					// Push the tool output to the context and re-compute with the tool output
					ctx.Push(types.Message{Role: types.RoleTool, Content: cm.ToList([]types.MessageContent{types.NewMessageContent(*toolResult)})})
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
			// overwrite the final message with the last message
			finalMsg = msg
			// Write full message to the output stream
			bytes, err := json.Marshal(finalMsg)
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
