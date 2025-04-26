package mcp

import (
	"context"
	"encoding/json" // Import encoding/json
	"fmt"
	"log"

	"masax-mcp/internal/masax" // Import masax client package

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	masaClient *masax.Client // Add Masa X client
}

// NewServer creates and configures a new MCP server instance, accepting the masax client.
func NewServer(client *masax.Client) (*MCPServer, error) {
	if client == nil {
		return nil, fmt.Errorf("masax client cannot be nil")
	}
	s := server.NewMCPServer(serverName, serverVersion)

	mcpServer := &MCPServer{
		MCPServer:  s,
		masaClient: client, // Store the client
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
		// Add max_results argument (using WithNumber)
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of search results to return (optional)"),
			// Add constraints if needed, e.g., mcp.Min(1)
		),
	)

	s.AddTool(searchTool, s.handleMasaXSearch)

	// Define the Masa X Search Result Resource (dynamic)
	searchResultResource := mcp.NewResource(
		searchResultResourcePrefix+"{"+searchIDParam+"}",
		"MasaX Search Result",
		mcp.WithResourceDescription("Represents the results of a specific Masa X API search."),
		mcp.WithMIMEType(jsonMimeType),
	)

	s.AddResource(searchResultResource, s.handleReadSearchResult)

	return nil
}

// handleMasaXSearch uses mcp.CallToolRequest and now returns the result content directly.
func (s *MCPServer) handleMasaXSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing or invalid 'query' argument"), nil
	}

	// Extract optional max_results (default to 0 or a reasonable value if needed)
	maxResults := 0 // Default to no limit specified or handle as per API needs
	if val, exists := request.Params.Arguments["max_results"]; exists {
		if num, ok := val.(float64); ok { // JSON numbers often decode as float64
			maxResults = int(num)
		}
	}

	fmt.Printf("Received search request for query: '%s', max_results: %d\n", query, maxResults)

	// 1. Call the actual Masa X API using s.masaClient
	searchResponse, err := s.masaClient.Search(ctx, query, maxResults)
	if err != nil {
		// Return API errors as tool errors for the LLM
		errMsg := fmt.Sprintf("Masa X API error: %v", err)
		log.Println(errMsg) // Log the error server-side too
		return mcp.NewToolResultError(errMsg), nil
	}

	// 2. Marshal the successful response to JSON
	jsonData, err := json.MarshalIndent(searchResponse, "", "  ") // Use MarshalIndent for readability
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal Masa X response: %v", err)
		log.Println(errMsg)
		return mcp.NewToolResultError(errMsg), nil // Internal server error
	}

	// 3. Generate a unique search_id if needed for the resource URI.
	//    For simplicity, let's just use the query for now, but UUID or hash is better.
	searchID := query // Simplistic ID
	resultURI := searchResultResourcePrefix + searchID

	// 4. Construct the resource content that the tool will return
	resultContents := mcp.TextResourceContents{
		URI:      resultURI, // URI representing this specific result
		MIMEType: jsonMimeType,
		Text:     string(jsonData), // The actual JSON string from API
	}

	// 5. Return the result using NewToolResultResource, embedding the content
	return mcp.NewToolResultResource(
		fmt.Sprintf("Masa X search results for query: '%s'", query),
		resultContents,
	), nil
}

// handleReadSearchResult uses mcp.ReadResourceRequest and returns []mcp.ResourceContents.
// This handler might become less relevant if the tool always returns full results.
// For now, it simulates fetching based on ID (which is just the query in this simple version).
func (s *MCPServer) handleReadSearchResult(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	searchID, ok := request.Params.Arguments[searchIDParam].(string) // ID is passed via arguments
	if !ok || searchID == "" {
		return nil, fmt.Errorf("missing '%s' argument in resource request for URI %s", searchIDParam, request.Params.URI)
	}

	fmt.Printf("Received request to read search results for id/query: %s\n", searchID)

	// Simulate re-fetching based on the ID (which is the query here)
	// In a real scenario, might query a cache or re-run the search
	searchResponse, err := s.masaClient.Search(ctx, searchID, 0) // Assume default maxResults for direct fetch
	if err != nil {
		// Return API errors - Resource not found might be appropriate here too
		errMsg := fmt.Sprintf("Failed to retrieve results for id '%s': %v", searchID, err)
		log.Println(errMsg)
		// Consider returning MCP error RESOURCE_NOT_FOUND if applicable
		// For now, just return nil content, error indicates failure
		return nil, fmt.Errorf(errMsg)
	}

	// Marshal the successful response to JSON
	jsonData, err := json.MarshalIndent(searchResponse, "", "  ")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal Masa X response for id '%s': %v", searchID, err)
		log.Println(errMsg)
		return nil, fmt.Errorf(errMsg) // Internal server error
	}

	// Return results as TextResourceContents with JSON MIME type
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI, // Use the requested URI
			MIMEType: jsonMimeType,
			Text:     string(jsonData), // The actual JSON string from API
		},
	}, nil
}
