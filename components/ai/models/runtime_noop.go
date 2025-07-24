//go:build !llama3

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
)

func runtime() export.Constructor {
	return func() (models.Format, error) {
		return nil, nil // No-op for non-llama3 builds
	}
}
