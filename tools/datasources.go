package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerDatasourceTools registers MCP tools for interacting with Grafana datasources.
func registerDatasourceTools(s *server.MCPServer, client *GrafanaClient) {
	registerListDatasources(s, client)
	registerGetDatasourceByName(s, client)
	registerGetDatasourceByUID(s, client)
}

// registerListDatasources registers a tool to list all configured datasources.
func registerListDatasources(s *server.MCPServer, client *GrafanaClient) {
	s.AddTool(
		mcp.NewTool(
			"list_datasources",
			mcp.WithDescription("List all datasources configured in Grafana"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := datasources.NewGetAllDataSourcesParams()
			resp, err := client.Datasources.GetAllDataSources(params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list datasources: %s", err)), nil
			}

			result, err := json.Marshal(resp.Payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal datasources: %s", err)), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	)
}

// registerGetDatasourceByName registers a tool to retrieve a specific datasource by name.
func registerGetDatasourceByName(s *server.MCPServer, client *GrafanaClient) {
	s.AddTool(
		mcp.NewTool(
			"get_datasource_by_name",
			mcp.WithDescription("Get a datasource by its name"),
			mcp.WithString(
				"name",
				mcp.Required(),
				mcp.Description("The name of the datasource to retrieve"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, ok := req.Params.Arguments["name"].(string)
			if !ok || name == "" {
				return mcp.NewToolResultError("name is required"), nil
			}

			params := datasources.NewGetDataSourceByNameParams().WithName(name)
			resp, err := client.Datasources.GetDataSourceByName(params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get datasource %q: %s", name, err)), nil
			}

			result, err := json.Marshal(resp.Payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal datasource: %s", err)), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	)
}

// registerGetDatasourceByUID registers a tool to retrieve a specific datasource by UID.
// Added this because I often know the UID from dashboard JSON and it's faster than listing all.
func registerGetDatasourceByUID(s *server.MCPServer, client *GrafanaClient) {
	s.AddTool(
		mcp.NewTool(
			"get_datasource_by_uid",
			mcp.WithDescription("Get a datasource by its UID"),
			mcp.WithString(
				"uid",
				mcp.Required(),
				mcp.Description("The UID of the datasource to retrieve"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uid, ok := req.Params.Arguments["uid"].(string)
			if !ok || uid == "" {
				return mcp.NewToolResultError("uid is required"), nil
			}

			params := datasources.NewGetDataSourceByUIDParams().WithUID(uid)
			resp, err := client.Datasources.GetDataSourceByUID(params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get datasource with uid %q: %s", uid, err)), nil
			}

			result, err := json.Marshal(resp.Payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal datasource: %s", err)), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	)
}
