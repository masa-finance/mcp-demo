package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	// "masax-mcp/internal/masax" // Import needed later
)

// Define constants for server name, version, and resource/tool names
const (
	serverName                 = "MasaX_MCP_Server"
	serverVersion              = "0.1.0"
	searchToolName             = "masa_x_search"
	searchResultResourcePrefix = "masax://search/results/"
	searchIDParam              = "search_id" // Consistent param name
	jsonMimeType               = "application/json"
)

// MCPServer wraps the mcp-go server implementation.
type MCPServer struct {
	*server.MCPServer
	// masaClient *masax.Client // Add Masa X client later
}

// NewServer creates and configures a new MCP server instance.
func NewServer() (*MCPServer, error) {
	s := server.NewMCPServer(serverName, serverVersion)

	// TODO: Initialize Masa X client
	// masaClient := masax.NewClient(/* config */)

	mcpServer := &MCPServer{
		MCPServer: s,
		// masaClient: masaClient,
	}

	if err := mcpServer.registerComponents(); err != nil {
		return nil, fmt.Errorf("failed to register MCP components: %w", err)
	}

	return mcpServer, nil
}

// registerComponents defines and registers MCP tools and resources.
func (s *MCPServer) registerComponents() error {
	// Define the Masa X Search Tool using README patterns
	searchTool := mcp.NewTool(
		searchToolName,
		mcp.WithDescription("Performs a search using the Masa X API and returns the results."),
		mcp.WithString(
			"query",
			mcp.Description("The search query string."),
			mcp.Required(),
		),
		// Add other potential arguments like count, specific user, etc.
		// mcp.WithInteger("count", mcp.Description("Number of results"), ...),
	)

	// Add the handler for the Search Tool (handler signature should use mcp.CallToolRequest)
	s.AddTool(searchTool, s.handleMasaXSearch)

	// Define the Masa X Search Result Resource (dynamic)
	searchResultResource := mcp.NewResource(
		searchResultResourcePrefix+"{"+searchIDParam+"}",
		"MasaX Search Result",
		mcp.WithResourceDescription("Represents the results of a specific Masa X API search."),
		mcp.WithMIMEType(jsonMimeType),
	)

	// Add the handler for the Search Result Resource (handler signature uses mcp.ReadResourceRequest)
	s.AddResource(searchResultResource, s.handleReadSearchResult)

	return nil
}

// handleMasaXSearch uses mcp.CallToolRequest and now returns the result content directly.
func (s *MCPServer) handleMasaXSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing or invalid 'query' argument"), nil
	}

	fmt.Printf("Received search request for query: %s\n", query)

	// TODO:
	// 1. Call the Masa X API using s.masaClient.Search(query, ...)
	// 2. Generate a unique search_id if needed for the resource URI.
	searchID := "temp_search_id_for_" + query // Placeholder ID generation
	resultURI := searchResultResourcePrefix + searchID

	// Simulate fetching/generating the result content here for now
	placeholderData := fmt.Sprintf(`{"searchId": "%s", "query": "%s", "results": ["result1 for query", "result2 for query"]}`, searchID, query)

	// Construct the resource content that the tool will return
	resultContents := mcp.TextResourceContents{
		URI:      resultURI, // URI representing this specific result
		MIMEType: jsonMimeType,
		Text:     placeholderData, // The actual JSON string
	}

	// Return the result using NewToolResultResource, embedding the content
	return mcp.NewToolResultResource(
		fmt.Sprintf("Masa X search results for query: %s", query), // Simple text description
		resultContents, // The actual resource content
	), nil
}

// handleReadSearchResult uses mcp.ReadResourceRequest and returns []mcp.ResourceContents.
func (s *MCPServer) handleReadSearchResult(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	searchID, ok := request.Params.Arguments[searchIDParam].(string)
	if !ok || searchID == "" {
		// Access URI via request.Params.URI
		return nil, fmt.Errorf("missing '%s' parameter in resource URI %s", searchIDParam, request.Params.URI)
	}

	fmt.Printf("Received request to read search results for id: %s\n", searchID)

	// TODO:
	// 1. Retrieve the actual search results associated with searchID (e.g., from a cache or by re-querying API if necessary)
	// 2. Format the results into the JSON string.

	// Placeholder response (could fetch from cache or re-run query based on searchID)
	placeholderData := fmt.Sprintf(`{"searchId": "%s", "results": ["result1 for id", "result2 for id"]}`, searchID)

	// Return results as TextResourceContents with JSON MIME type
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI, // Use the requested URI
			MIMEType: jsonMimeType,
			Text:     placeholderData, // The actual JSON string
		},
	}, nil
}
