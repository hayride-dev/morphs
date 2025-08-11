//go:build !llama_3_1 && !qwen_2_5 && !qwen_3 && !gpt_oss

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
)

func build() export.Constructor {
	return func() (models.Format, error) {
		return nil, nil // No-op for non-llama3 builds
	}
}
