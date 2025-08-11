package main

import (
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents/export"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

var _ agents.Agent = (*defaultAgent)(nil)

type defaultAgent struct {
	name        string
	instruction string
	format      models.Format
	graph       graph.GraphExecutionContextStream

	// Tools and Context are optional
	tools   tools.Tools
	context ctx.Context
}

func init() {
	export.Agent(constructor)
}

func constructor(name string, instruction string, format models.Format, graph graph.GraphExecutionContextStream, tools tools.Tools, context ctx.Context) (agents.Agent, error) {
	if format == nil {
		return nil, fmt.Errorf("format is required for agent")
	}

	if graph == nil {
		return nil, fmt.Errorf("graph is required for agent")
	}

	agent := &defaultAgent{
		name:        name,
		instruction: instruction,
		tools:       tools,
		context:     context,
		format:      format,
		graph:       graph,
	}

	// If context is set, push the initial instruction message
	if context != nil {
		content := []types.MessageContent{}
		content = append(content, types.NewMessageContent(types.Text(instruction)))

		// If tools are set, list them and append to content
		if tools != nil {
			result, err := tools.List("")
			if err != nil {
				return nil, err
			}

			if result.Tools.Len() > 0 {
				// Append the list of tools to the content
				content = append(content, types.NewMessageContent(result.Tools))
			}
		}

		// Push message to the context
		msg := types.Message{Role: types.RoleSystem, Content: cm.ToList(content)}
		agent.context.Push(cm.Reinterpret[types.Message](msg))
	}

	return agent, nil
}

func (a *defaultAgent) Name() string {
	return a.name
}

func (a *defaultAgent) Instruction() string {
	return a.instruction
}

func (a *defaultAgent) Capabilities() ([]types.Tool, error) {
	if a.tools == nil {
		return nil, fmt.Errorf("tools are not set for agent %s", a.name)
	}

	result, err := a.tools.List("")
	if err != nil {
		return nil, err
	}

	return result.Tools.Slice(), nil
}

func (a *defaultAgent) Context() ([]types.Message, error) {
	if a.context == nil {
		return nil, fmt.Errorf("context is not set for agent %s", a.name)
	}

	msgs, err := a.context.Messages()
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (a *defaultAgent) Compute(message types.Message) (*types.Message, error) {
	var msgs []types.Message
	// Push message to context
	if a.context != nil {
		if err := a.context.Push(message); err != nil {
			return nil, fmt.Errorf("failed to push message to context: %w", err)
		}
		// Get all context messages
		m, err := a.context.Messages()
		if err != nil {
			return nil, fmt.Errorf("failed to get context messages: %w", err)
		}

		msgs = m
	} else {
		msgs = []types.Message{message}
	}

	// Format encode the messages
	data, err := a.format.Encode(msgs...)
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
	namedTensorStream, err := a.graph.Compute(inputs)
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

		// TODO:: Optionally write RAW output to a writer
		// to get the raw output back faster, but would require an updated interface for agent compute
	}

	// Decode Message
	msg, err := a.format.Decode(text) // Gets us our API Message
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	// Push to Context if set
	if a.context != nil {
		if err := a.context.Push(*msg); err != nil {
			return nil, fmt.Errorf("failed to push message to context: %w", err)
		}
	}

	// Return the final message
	return msg, nil
}

func (a *defaultAgent) Execute(params types.CallToolParams) (*types.CallToolResult, error) {
	if a.tools == nil {
		return nil, fmt.Errorf("tools are not set for agent %s", a.name)
	}

	result, err := a.tools.Call(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func main() {}
