package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents/export"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

var _ agents.Agent = (*defaultAgent)(nil)

type defaultAgent struct {
	name        string
	instruction string

	// Tools and Context are optional
	tools   tools.Tools
	context ctx.Context
}

func init() {
	export.Agent(constructor)
}

func constructor(name string, instruction string, tools tools.Tools, context ctx.Context) (agents.Agent, error) {

	agent := &defaultAgent{
		name:        name,
		instruction: instruction,
		tools:       tools,
		context:     context,
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

func (a *defaultAgent) Push(msg types.Message) error {
	if a.context == nil {
		return fmt.Errorf("context is not set for agent %s", a.name)
	}
	return a.context.Push(cm.Reinterpret[types.Message](msg))
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
