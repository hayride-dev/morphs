package main

import (
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/mcp/tools/export"
	"github.com/hayride-dev/bindings/go/hayride/types"
)

var _ tools.Tools = (*noopTools)(nil)

type noopTools struct{}

func (n *noopTools) Call(params types.CallToolParams) (*types.CallToolResult, error) {
	return &types.CallToolResult{}, nil
}

func (n *noopTools) List(cursor string) (*types.ListToolsResult, error) {
	return &types.ListToolsResult{}, nil
}

func constructor() (tools.Tools, error) {
	return &noopTools{}, nil
}

func init() {
	export.Tools(constructor)
}

func main() {}
