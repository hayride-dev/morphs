package main

import (
	"fmt"

	"github.com/hayride-dev/bindings/go/hayride/mcp/tools"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server"
	"github.com/hayride-dev/bindings/go/hayride/x/net/http/server/export"
	mcpserver "github.com/hayride-dev/bindings/go/mcp/server"
)

func init() {
	fmt.Println("Initializing MCP server with tools")

	// Get tools from import and add them to the server
	toolbox, err := tools.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to create tools: %v", err))
	}

	// To include authentication import and add auth provider to the server
	/*
		provider, err := auth.New()
		if err != nil {
			panic(fmt.Sprintf("Failed to create auth provider: %v", err))
		}
	*/

	// Create the MCP server router with the tools and auth provider
	mux, err := mcpserver.NewMCPRouter(
		mcpserver.WithName("MCP Server"),
		mcpserver.WithVersion("0.0.1"),
		mcpserver.WithTools(toolbox),
		// server.WithAuthProvider(provider),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create MCP server: %v", err))
	}

	export.ServerConfig(mux, server.ServerConfig{
		Address: "http://localhost:8083",
	})
}

func main() {

}
