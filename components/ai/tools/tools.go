package main

import (
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/tools/internal/gen/hayride/ai/tools"
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
	tools.Exports.Tools.Call = call
	tools.Exports.Tools.Capabilities = capabilities
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

func call(self cm.Rep, input tools.ToolInput) (result cm.Result[tools.ToolOutputShape, tools.ToolOutput, tools.ErrorCode]) {
	return cm.OK[cm.Result[tools.ToolOutputShape, tools.ToolOutput, tools.ErrorCode]](tools.ToolOutput{})
}

func capabilities(self cm.Rep) (result cm.Result[cm.List[tools.ToolSchema], cm.List[tools.ToolSchema], tools.Error]) {
	return cm.OK[cm.Result[cm.List[tools.ToolSchema], cm.List[tools.ToolSchema], tools.Error]](cm.List[tools.ToolSchema]{})
}

func destructor(self cm.Rep) {
	delete(resourceTable.tools, self)
}

func main() {}
