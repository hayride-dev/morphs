//go:build ory

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/mcp/auth/export"
	"github.com/hayride-dev/morphs/components/mcp/auth/ory"
)

func build() export.Constructor {
	return ory.Constructor
}
