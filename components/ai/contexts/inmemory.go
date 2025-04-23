//go:build in_memory

package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/exports/ai/ctx"
	"github.com/hayride-dev/bindings/go/gen/types/hayride/ai/types"
)

var _ ctx.Context = (*inMemoryContext)(nil)

type inMemoryContext struct {
	context []types.Message
}

func init() {
	c := &inMemoryContext{
		context: make([]types.Message, 0),
	}
	ctx.Export(c)
}

func (c *inMemoryContext) Push(messages ...types.Message) error {
	for _, m := range messages {
		c.context = append(c.context, m)
	}
	return nil
}

func (c *inMemoryContext) Messages() ([]types.Message, error) {
	return c.context, nil
}

func (c *inMemoryContext) Next() (types.Message, error) {
	if len(c.context) == 0 {
		return types.Message{}, fmt.Errorf("missing messages")
	}
	return c.context[len(c.context)-1], nil
}

func main() {}
