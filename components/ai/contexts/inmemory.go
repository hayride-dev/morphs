//go:build in_memory

package main

import (
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/contexts/internal/gen/hayride/ai/context"
	"github.com/hayride-dev/morphs/components/ai/contexts/internal/gen/hayride/ai/types"
	"go.bytecodealliance.org/cm"
)

type resources struct {
	ctx map[context.Context]*inMemoryContext
}

type inMemoryContext struct {
	context []types.Message
}

func (c *inMemoryContext) push(msg context.Message) error {
	c.context = append(c.context, msg)
	return nil
}

func (c *inMemoryContext) messages() []context.Message {
	return c.context
}

var resourceTable = resources{
	ctx: make(map[context.Context]*inMemoryContext),
}

func init() {
	context.Exports.Context.Constructor = constructor
	context.Exports.Context.Push = push
	context.Exports.Context.Messages = messages
}

func constructor() context.Context {
	ctx := &inMemoryContext{
		context: make([]types.Message, 0),
	}

	context.Exports.Context.Push = push
	context.Exports.Context.Messages = messages

	v := context.ContextResourceNew((cm.Rep(uintptr(unsafe.Pointer(ctx)))))
	resourceTable.ctx[v] = ctx
	return v
}

func push(self cm.Rep, msg context.Message) cm.Result[context.Error, struct{}, context.Error] {
	ctx, ok := resourceTable.ctx[context.Context(self)]
	if !ok {
		wasiErr := context.ErrorResourceNew(cm.Rep(context.ErrorCodePushError))
		return cm.Err[cm.Result[context.Error, struct{}, context.Error]](wasiErr)
	}

	if err := ctx.push(msg); err != nil {
		wasiErr := context.ErrorResourceNew(cm.Rep(context.ErrorCodePushError))
		return cm.Err[cm.Result[context.Error, struct{}, context.Error]](wasiErr)
	}
	return cm.Result[context.Error, struct{}, context.Error]{}
}

func messages(self cm.Rep) (result cm.Result[cm.List[context.Message], cm.List[context.Message], context.Error]) {
	ctx, ok := resourceTable.ctx[context.Context(self)]
	if !ok {
		wasiErr := context.ErrorResourceNew(cm.Rep(context.ErrorCodeMessageNotFound))
		return cm.Err[cm.Result[cm.List[context.Message], cm.List[context.Message], context.Error]](wasiErr)
	}

	return cm.OK[cm.Result[cm.List[context.Message], cm.List[context.Message], context.Error]](cm.ToList(ctx.messages()))
}

func main() {}
