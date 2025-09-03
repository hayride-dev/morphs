//go:build in_memory

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx"
	"github.com/hayride-dev/bindings/go/hayride/ai/ctx/export"
)

var _ ctx.Context = (*inMemoryContext)(nil)

type inMemoryContext struct {
	context []ai.Message
}

func (c *inMemoryContext) Push(msg ...ai.Message) error {
	c.context = append(c.context, msg...)
	return nil
}

func (c *inMemoryContext) Messages() ([]ai.Message, error) {
	return c.context, nil
}

func constructor() (ctx.Context, error) {
	return &inMemoryContext{
		context: make([]ai.Message, 0),
	}, nil
}

func init() {
	export.Context(constructor)
}

func main() {}
