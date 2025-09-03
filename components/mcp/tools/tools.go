package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/hayride/mcp"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools/export"
	"github.com/hayride-dev/morphs/components/util/datetime/pkg/datetime"
	"go.bytecodealliance.org/cm"
)

var _ tools.Tools = (*defaultTools)(nil)

type defaultTools struct{}

func (n *defaultTools) Call(params mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if params.Name != "datetime" {
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}

	// Call our imported datetime tool
	date := datetime.Date()

	content := mcp.NewContent(mcp.TextContent{
		ContentType: "text",
		Text:        date,
	})

	return &mcp.CallToolResult{
		Content: cm.ToList([]mcp.Content{content}),
	}, nil
}

func (n *defaultTools) List(cursor string) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{
		Tools: cm.ToList([]mcp.Tool{
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
