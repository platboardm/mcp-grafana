// Package tools provides MCP tool implementations for interacting with Grafana.
// Each tool corresponds to a specific Grafana API capability exposed via the
// Model Context Protocol (MCP).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GrafanaClient holds configuration for communicating with a Grafana instance.
type GrafanaClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewGrafanaClient creates a new GrafanaClient with the provided base URL and API key.
// Uses a 30-second timeout as a reasonable default for most Grafana instances.
// Increase this value if you're working with particularly large dashboards or
// slow remote instances.
func NewGrafanaClient(baseURL, apiKey string) *GrafanaClient {
	return &GrafanaClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// get performs an authenticated GET request against the Grafana API.
func (c *GrafanaClient) get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.HTTPClient.Do(req)
}

// RegisterTools registers all available Grafana MCP tools with the given server.
func RegisterTools(s *server.MCPServer, client *GrafanaClient) {
	registerSearchDashboards(s, client)
	registerGetDashboard(s, client)
	registerListDataSources(s, client)
}

// registerSearchDashboards registers the search_dashboards tool.
func registerSearchDashboards(s *server.MCPServer, client *GrafanaClient) {
	s.AddTool(
		mcp.NewTool("search_dashboards",
			mcp.WithDescription("Search for dashboards in Grafana by query string or tag."),
			mcp.WithString("query",
				mcp.Description("Search query string to filter dashboards by title."),
			),
			mcp.WithString("tag",
				mcp.Description("Filter dashboards by tag."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := url.Values{}
			params.Set("type", "dash-db")

			if q, ok := req.Params.Arguments["query"].(string); ok && q != "" {
				params.Set("query", q)
			}
			if tag, ok := req.Params.Arguments["tag"].(string); ok && tag != "" {
				params.Set("tag", tag)
			}

			resp, err := client.get(ctx, "/api/search", params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("searching dashboards: %v", err)), nil
			}
			defer resp.Body.Close()

			var result interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return mcp.NewToolResultError(fmt.Sp