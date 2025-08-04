//go:build qwen_3

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
	"github.com/hayride-dev/morphs/components/ai/models/qwen"
)

func build() export.Constructor {
	return qwen.ConstructorQwen_3
}
