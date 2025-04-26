package main

import (
	"log"
	"os" // Import os package

	"masax-mcp/internal/masax" // Import masax client package
	"masax-mcp/internal/mcp"

	"github.com/joho/godotenv"           // Import godotenv
	"github.com/mark3labs/mcp-go/server" // Import server package
)

func main() {
	// Load .env file. Handle errors, but maybe continue if not found?
	err := godotenv.Load() // Load .env from current directory
	if err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("MASA_API_KEY")
	if apiKey == "" {
		log.Fatalf("Error: MASA_API_KEY environment variable not set.")
	}

	// Create Masa X client
	masaClient, err := masax.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create Masa X client: %v", err)
	}

	// Initialize MCP server, passing the client
	mcpServer, err := mcp.NewServer(masaClient)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start the server using the server package function
	if err := server.ServeStdio(mcpServer.MCPServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
