//go:build llama_3_1

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
	"github.com/hayride-dev/morphs/components/ai/models/llama3"
)

func build() export.Constructor {
	return llama3.Constructor_3_1
}
