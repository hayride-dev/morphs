//go:build llama3

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
	"github.com/hayride-dev/morphs/components/ai/models/llama3"
)

func runtime() export.Constructor {
	return llama3.Constructor
}
