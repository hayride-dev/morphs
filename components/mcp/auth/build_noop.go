//go:build !ory

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/mcp/auth"
	"github.com/hayride-dev/bindings/go/hayride/mcp/auth/export"
)

func build() export.Constructor {
	return func() (auth.Provider, error) {
		return nil, nil // No-op for non-llama3 builds
	}
}
