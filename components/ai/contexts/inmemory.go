//go:build in_memory

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx/export"
	"github.com/hayride-dev/bindings/go/hayride/types"
)

var _ ctx.Context = (*inMemoryContext)(nil)

type inMemoryContext struct {
	context []types.Message
}

func (c *inMemoryContext) Push(msg ...types.Message) error {
	c.context = append(c.context, msg...)
	return nil
}

func (c *inMemoryContext) Messages() ([]types.Message, error) {
	return c.context, nil
}

func constructor() (ctx.Context, error) {
	return &inMemoryContext{
		context: make([]types.Message, 0),
	}, nil
}

func init() {
	export.Context(constructor)
}

func main() {}
