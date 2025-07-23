package main

import (
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
	tools       tools.Tools
	context     ctx.Context
	format      models.Format
	graph       graph.GraphExecutionContextStream
}

func init() {
	export.Agent(constructor)
}

func constructor(name string, instruction string, tools tools.Tools, context ctx.Context, format models.Format, graph graph.GraphExecutionContextStream) (agents.Agent, error) {
	agent := &defaultAgent{
		name:        name,
		instruction: instruction,
		tools:       tools,
		context:     context,
		format:      format,
		graph:       graph,
	}

	content := []types.MessageContent{}
	content = append(content, types.NewMessageContent(types.Text(instruction)))

	result, err := tools.List("")
	if err != nil {
		return nil, err
	}

	// Append the list of tools to the content
	content = append(content, types.NewMessageContent(result.Tools))

	// Push message to the context
	msg := types.Message{Role: 1, Content: cm.ToList(content)}
	agent.context.Push(cm.Reinterpret[types.Message](msg))

	return agent, nil
}

func (a *defaultAgent) Name() string {
	return a.name
}

func (a *defaultAgent) Instruction() string {
	return a.instruction
}

func (a *defaultAgent) Tools() tools.Tools {
	return a.tools
}

func (a *defaultAgent) Context() ctx.Context {
	return a.context
}

func (a *defaultAgent) Format() models.Format {
	return a.format
}

func (a *defaultAgent) Graph() graph.GraphExecutionContextStream {
	return a.graph
}

func main() {}
