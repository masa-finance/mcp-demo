package masax

import (
	"bytes" // Added for request body
	"context"
	"encoding/json" // Added for JSON marshaling/unmarshaling
	"fmt"
	"io" // Added for reading response body
	"net/http"
	"net/url" // Added for joining URL paths
	"time"
	// "os" // No longer needed directly here
)

// --- Request Structures ---

// SearchRequest represents the JSON body sent to the Masa X Search API.
type SearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

// --- Response Structures ---

// PublicMetrics holds engagement data for a search result item.
type PublicMetrics struct {
	RetweetCount int `json:"retweet_count"`
	ReplyCount   int `json:"reply_count"`
	LikeCount    int `json:"like_count"`
	QuoteCount   int `json:"quote_count"`
}

// SearchResult represents a single item returned by the Masa X Search API.
type SearchResult struct {
	ID            string        `json:"id"`
	Text          string        `json:"text"`
	AuthorID      string        `json:"author_id"`
	CreatedAt     time.Time     `json:"created_at"`
	PublicMetrics PublicMetrics `json:"public_metrics"`
	URL           string        `json:"url"`
}

// SearchMetadata contains pagination or summary info for the search response.
type SearchMetadata struct {
	TotalResults int    `json:"total_results"`
	NextToken    string `json:"next_token,omitempty"`
}

// SearchResponse represents the overall successful response from the Masa X Search API.
type SearchResponse struct {
	Items    []SearchResult `json:"items"`
	Metadata SearchMetadata `json:"metadata"`
}

// ErrorDetail represents the structure within an API error response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response from the Masa X Search API.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// --- Client Implementation ---

const (
	defaultBaseURL = "https://data.dev.masalabs.ai/api/v1"
	searchPath     = "/search/live/twitter"
)

// Client manages communication with the Masa X API.
type Client struct {
	httpClient *http.Client
	apiBaseURL string
	apiKey     string
}

// NewClient creates a new Masa X API client.
func NewClient(apiKey string, options ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("masa X API key is required")
	}
	c := &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		apiBaseURL: defaultBaseURL,
		apiKey:     apiKey,
	}
	for _, opt := range options {
		opt(c)
	}
	return c, nil
}

// ClientOption defines a functional option for configuring the Client.
type ClientOption func(*Client)

// WithHTTPClient allows providing a custom http.Client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithBaseURL allows overriding the default API base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		if baseURL != "" {
			c.apiBaseURL = baseURL
		}
	}
}

// Search performs a search query against the Masa X API.
func (c *Client) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
	// 1. Create SearchRequest and marshal to JSON
	searchReq := SearchRequest{
		Query:      query,
		MaxResults: maxResults,
	}
	reqBodyBytes, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 2. Construct URL and create request
	// Use url.JoinPath for safer path joining (requires Go 1.19+)
	fullURL, err := url.JoinPath(c.apiBaseURL, searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create search URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 3. Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 4. Send request
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	// 5. Read response body
	respBodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 6. Check status code and handle errors
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		var apiError ErrorResponse
		if json.Unmarshal(respBodyBytes, &apiError) == nil && apiError.Error.Message != "" {
			// Return structured API error
			return nil, fmt.Errorf("masa X API error (HTTP %d - %s): %s", httpResp.StatusCode, apiError.Error.Code, apiError.Error.Message)
		}
		// Return generic HTTP error if body parsing failed or error format unexpected
		return nil, fmt.Errorf("masa X API request failed with HTTP status %d: %s", httpResp.StatusCode, string(respBodyBytes))
	}

	// 7. Unmarshal successful response
	var searchResp SearchResponse
	if err := json.Unmarshal(respBodyBytes, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal successful response body: %w", err)
	}

	return &searchResp, nil
}
