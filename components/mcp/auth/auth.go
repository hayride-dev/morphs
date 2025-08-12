package main

import "github.com/hayride-dev/bindings/go/hayride/mcp/auth/export"

func init() {
	export.Provider(build())
}

func main() {}
