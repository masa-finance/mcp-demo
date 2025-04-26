package main

import (
	"log"

	"masax-mcp/internal/mcp"

	"github.com/mark3labs/mcp-go/server" // Import server package
)

func main() {
	// Initialize and run the MCP server
	mcpServer, err := mcp.NewServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start the server using the server package function
	// Pass the underlying *server.MCPServer instance
	if err := server.ServeStdio(mcpServer.MCPServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
