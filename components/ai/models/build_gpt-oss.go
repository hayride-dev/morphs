//go:build gpt_oss

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
	"github.com/hayride-dev/morphs/components/ai/models/gpt"
)

func build() export.Constructor {
	return gpt.ConstructorGPTOSS
}
