package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerDashboardTools registers all dashboard-related MCP tools with the server.
func registerDashboardTools(s *server.MCPServer, c *GrafanaClient) {
	registerGetDashboardByUID(s, c)
	registerListDashboards(s, c)
}

// registerGetDashboardByUID registers a tool to fetch a dashboard by its UID.
func registerGetDashboardByUID(s *server.MCPServer, c *GrafanaClient) {
	tool := mcp.NewTool(
		"get_dashboard_by_uid",
		mcp.WithDescription("Retrieve a Grafana dashboard by its UID. Returns the full dashboard JSON model along with metadata."),
		mcp.WithString(
			"uid",
			mcp.Required(),
			mcp.Description("The unique identifier (UID) of the dashboard to retrieve."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid, ok := req.Params.Arguments["uid"].(string)
		if !ok || uid == "" {
			return mcp.NewToolResultError("uid is required and must be a non-empty string"), nil
		}

		params := dashboards.NewGetDashboardByUIDParams().WithUID(uid)
		resp, err := c.Dashboards.GetDashboardByUID(params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get dashboard by UID %q: %s", uid, err)), nil
		}

		result, err := json.MarshalIndent(resp.Payload, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal dashboard response: %s", err)), nil
		}

		return mcp.NewToolResultText(string(result)), nil
	})
}

// registerListDashboards registers a tool to list dashboards with optional filtering.
func registerListDashboards(s *server.MCPServer, c *GrafanaClient) {
	tool := mcp.NewTool(
		"list_dashboards",
		mcp.WithDescription("List Grafana dashboards, optionally filtered by folder UID or a search query."),
		mcp.WithString(
			"query",
			mcp.Description("Optional search query to filter dashboards by title."),
		),
		mcp.WithString(
			"folder_uid",
			mcp.Description("Optional folder UID to restrict the listing to dashboards within that folder."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, _ := req.Params.Arguments["query"].(string)
		folderUID, _ := req.Params.Arguments["folder_uid"].(string)

		// Reuse the existing search implementation with type=dash-db
		dashType := "dash-db"
		results, err := searchDashboards(ctx, c, query, folderUID, dashType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list dashboards: %s", err)), nil
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("No dashboards found."), nil
		}

		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal dashboards list: %s", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	})
}
