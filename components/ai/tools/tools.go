package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools/export"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"github.com/hayride-dev/morphs/components/ai/tools/datetime/pkg/datetime"
	"go.bytecodealliance.org/cm"
)

var _ tools.Tools = (*defaultTools)(nil)

type defaultTools struct{}

func (n *defaultTools) Call(params types.CallToolParams) (*types.CallToolResult, error) {
	if params.Name != "datetime" {
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}

	// Call our imported datetime tool
	date := datetime.Date()

	content := types.NewContent(types.TextContent{
		ContentType: "text",
		Text:        date,
	})

	return &types.CallToolResult{
		Content: cm.ToList([]types.Content{content}),
	}, nil
}

func (n *defaultTools) List(cursor string) (*types.ListToolsResult, error) {
	return &types.ListToolsResult{
		Tools: cm.ToList([]types.Tool{
			{
				Name:        "datetime",
				Title:       "Datetime",
				Description: "Provides the current date and time.",
			},
		}),
	}, nil
}

func constructor() (tools.Tools, error) {
	return &defaultTools{}, nil
}

func init() {
	export.Tools(constructor)
}

func main() {}
