//go:build in_memory

package main

import (
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/contexts/internal/gen/hayride/ai/context"
	"github.com/hayride-dev/morphs/components/ai/contexts/internal/gen/hayride/ai/types"
	"go.bytecodealliance.org/cm"
)

func init() {

}

type inMemoryContext struct {
	context []types.Message
}

func init() {
	context.Exports.Context.Constructor = constructor
}

func constructor() context.Context {
	ctx := &inMemoryContext{
		context: make([]types.Message, 0),
	}

	context.Exports.Context.Push = ctx.push
	context.Exports.Context.Messages = ctx.messages
	return context.ContextResourceNew((cm.Rep(uintptr(unsafe.Pointer(ctx)))))
}

func (c *inMemoryContext) push(self cm.Rep, msg context.Message) cm.Result[context.Error, struct{}, context.Error] {
	c.context = append(c.context, msg)
	return cm.Result[context.Error, struct{}, context.Error]{}
}

func (c *inMemoryContext) messages(self cm.Rep) (result cm.Result[cm.List[context.Message], cm.List[context.Message], context.Error]) {
	return cm.OK[cm.Result[cm.List[context.Message], cm.List[context.Message], context.Error]](cm.ToList(c.context))
}

func main() {}
