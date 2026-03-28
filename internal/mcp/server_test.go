package mcp

import (
	"context"
	"strings"
	"testing"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
)

// TestMCPEndToEnd tests the full MCP flow: initialize → list tools → call tools
// using the in-process transport (no network, but exercises the full mcp-go stack).
func TestMCPEndToEnd(t *testing.T) {
	svc := &mockAssetService{
		searchAssetsFunc: func(_ context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
			return []asset.SearchResult{
				{ID: "1", URN: "urn:bq:dataset.orders", Title: "orders", Type: "table", Service: "bigquery", Description: "Main orders table"},
				{ID: "2", URN: "urn:bq:dataset.customers", Title: "customers", Type: "table", Service: "bigquery", Description: "Customer records"},
			}, nil
		},
		getAssetByIDFunc: func(_ context.Context, id string) (asset.Asset, error) {
			return asset.Asset{
				ID:          "1",
				URN:         "urn:bq:dataset.orders",
				Name:        "orders",
				Type:        "table",
				Service:     "bigquery",
				Description: "Main orders table with all customer transactions",
				Owners:      []user.User{{Email: "alice@company.com"}, {Email: "bob@company.com"}},
				Data: map[string]interface{}{
					"columns": []interface{}{
						map[string]interface{}{"name": "order_id", "data_type": "INTEGER", "description": "Primary key"},
						map[string]interface{}{"name": "customer_id", "data_type": "INTEGER", "description": "FK to customers"},
						map[string]interface{}{"name": "amount", "data_type": "FLOAT", "description": "Order total in USD"},
					},
				},
			}, nil
		},
		getLineageFunc: func(_ context.Context, urn string, q asset.LineageQuery) (asset.Lineage, error) {
			return asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn:bq:raw_orders", Target: urn},
					{Source: urn, Target: "urn:bq:order_summary"},
				},
			}, nil
		},
		getTypesFunc: func(_ context.Context, _ asset.Filter) (map[asset.Type]int, error) {
			return map[asset.Type]int{
				"table":     42,
				"topic":     15,
				"dashboard": 8,
			}, nil
		},
		getAllAssetsFunc: func(_ context.Context, flt asset.Filter, _ bool) ([]asset.Asset, uint32, error) {
			return []asset.Asset{
				{ID: "1", URN: "urn:bq:dataset.orders", Name: "orders", Type: "table", Service: "bigquery"},
				{ID: "2", URN: "urn:bq:dataset.customers", Name: "customers", Type: "table", Service: "bigquery"},
			}, 100, nil
		},
	}

	srv := New(svc, namespace.DefaultNamespace)
	ctx := context.Background()

	// Create in-process MCP client (exercises the full mcp-go protocol stack)
	client, err := mcpclient.NewInProcessClient(srv.mcpServer)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}

	// Step 1: Initialize
	initResult, err := client.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	t.Logf("Server: %s v%s", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	if initResult.ServerInfo.Name != "compass" {
		t.Errorf("expected server name 'compass', got '%s'", initResult.ServerInfo.Name)
	}

	// Step 2: List tools
	tools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list tools failed: %v", err)
	}
	t.Logf("Available tools (%d):", len(tools.Tools))
	for _, tool := range tools.Tools {
		t.Logf("  - %s: %s", tool.Name, tool.Description)
	}

	expectedTools := []string{"search_assets", "get_asset", "get_lineage", "list_types", "get_all_assets"}
	if len(tools.Tools) != len(expectedTools) {
		t.Fatalf("expected %d tools, got %d", len(expectedTools), len(tools.Tools))
	}
	toolNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		toolNames[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}

	// Step 3: Call search_assets
	t.Run("search_assets", func(t *testing.T) {
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_assets",
				Arguments: map[string]any{"text": "orders"},
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		text := extractText(result)
		t.Logf("Result:\n%s", text)

		if !strings.Contains(text, "orders") {
			t.Error("expected result to contain 'orders'")
		}
		if !strings.Contains(text, "Found 2 assets") {
			t.Error("expected result to mention 2 assets found")
		}
	})

	// Step 4: Call get_asset
	t.Run("get_asset", func(t *testing.T) {
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_asset",
				Arguments: map[string]any{"id": "urn:bq:dataset.orders"},
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		text := extractText(result)
		t.Logf("Result:\n%s", text)

		if !strings.Contains(text, "orders") {
			t.Error("expected result to contain 'orders'")
		}
		if !strings.Contains(text, "alice@company.com") {
			t.Error("expected result to contain owner email")
		}
		if !strings.Contains(text, "order_id") {
			t.Error("expected result to contain column info")
		}
	})

	// Step 5: Call get_lineage
	t.Run("get_lineage", func(t *testing.T) {
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_lineage",
				Arguments: map[string]any{"urn": "urn:bq:dataset.orders"},
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		text := extractText(result)
		t.Logf("Result:\n%s", text)

		if !strings.Contains(text, "raw_orders") {
			t.Error("expected result to contain upstream 'raw_orders'")
		}
		if !strings.Contains(text, "order_summary") {
			t.Error("expected result to contain downstream 'order_summary'")
		}
	})

	// Step 6: Call list_types
	t.Run("list_types", func(t *testing.T) {
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_types",
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		text := extractText(result)
		t.Logf("Result:\n%s", text)

		if !strings.Contains(text, "table") {
			t.Error("expected result to contain 'table'")
		}
	})

	// Step 7: Call get_all_assets
	t.Run("get_all_assets", func(t *testing.T) {
		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_all_assets",
				Arguments: map[string]any{"size": 10},
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		text := extractText(result)
		t.Logf("Result:\n%s", text)

		if !strings.Contains(text, "Showing 2 of 100 assets") {
			t.Error("expected result to show pagination info")
		}
	})
}

func extractText(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
