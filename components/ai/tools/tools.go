package main

import (
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/tools/internal/gen/hayride/mcp/tools"
	"go.bytecodealliance.org/cm"
)

type resources struct {
	tools map[cm.Rep]*noopTools
}

var resourceTable = &resources{
	tools: make(map[cm.Rep]*noopTools),
}

func init() {
	tools.Exports.Tools.Constructor = constructor
	tools.Exports.Tools.CallTool = call
	tools.Exports.Tools.ListTools = list
	tools.Exports.Tools.Destructor = destructor
}

type noopTools struct{}

func constructor() tools.Tools {
	noop := &noopTools{}
	key := cm.Rep(uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&noop))))
	v := tools.ToolsResourceNew(key)
	resourceTable.tools[key] = noop
	return v
}

func call(self cm.Rep, params tools.CallToolParams) (result cm.Result[tools.CallToolResultShape, tools.CallToolResult, tools.Error]) {
	return cm.OK[cm.Result[tools.CallToolResultShape, tools.CallToolResult, tools.Error]](tools.CallToolResult{})
}

func list(self cm.Rep, cursor string) (result cm.Result[tools.ListToolsResultShape, tools.ListToolsResult, tools.Error]) {
	return cm.OK[cm.Result[tools.ListToolsResultShape, tools.ListToolsResult, tools.Error]](tools.ListToolsResult{})
}

func destructor(self cm.Rep) {
	delete(resourceTable.tools, self)
}

func main() {}
