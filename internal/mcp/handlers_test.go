package mcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
)

// mockAssetService is a test double for the AssetService interface.
type mockAssetService struct {
	searchAssetsFunc  func(ctx context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error)
	getAssetByIDFunc  func(ctx context.Context, id string) (asset.Asset, error)
	getLineageFunc    func(ctx context.Context, urn string, query asset.LineageQuery) (asset.Lineage, error)
	getTypesFunc      func(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error)
	getAllAssetsFunc   func(ctx context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error)
}

func (m *mockAssetService) SearchAssets(ctx context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
	if m.searchAssetsFunc != nil {
		return m.searchAssetsFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *mockAssetService) GetAssetByID(ctx context.Context, id string) (asset.Asset, error) {
	if m.getAssetByIDFunc != nil {
		return m.getAssetByIDFunc(ctx, id)
	}
	return asset.Asset{}, nil
}

func (m *mockAssetService) GetLineage(ctx context.Context, urn string, query asset.LineageQuery) (asset.Lineage, error) {
	if m.getLineageFunc != nil {
		return m.getLineageFunc(ctx, urn, query)
	}
	return asset.Lineage{}, nil
}

func (m *mockAssetService) GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error) {
	if m.getTypesFunc != nil {
		return m.getTypesFunc(ctx, flt)
	}
	return nil, nil
}

func (m *mockAssetService) GetAllAssets(ctx context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error) {
	if m.getAllAssetsFunc != nil {
		return m.getAllAssetsFunc(ctx, flt, withTotal)
	}
	return nil, 0, nil
}

func newTestServer(svc *mockAssetService) *Server {
	return New(svc, namespace.DefaultNamespace)
}

func callToolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestHandleSearchAssets(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when text is empty", func(t *testing.T) {
		s := newTestServer(&mockAssetService{})
		result, err := s.handleSearchAssets(ctx, callToolRequest(map[string]any{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("returns search results", func(t *testing.T) {
		svc := &mockAssetService{
			searchAssetsFunc: func(_ context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
				if cfg.Text != "orders" {
					t.Errorf("expected text 'orders', got '%s'", cfg.Text)
				}
				return []asset.SearchResult{
					{ID: "1", URN: "urn:bq:orders", Title: "orders", Type: "table", Service: "bigquery"},
				}, nil
			},
		}
		s := newTestServer(svc)
		result, err := s.handleSearchAssets(ctx, callToolRequest(map[string]any{"text": "orders"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Error("unexpected error result")
		}
		text := getTextContent(result)
		if text == "" {
			t.Error("expected non-empty text content")
		}
	})

	t.Run("passes filters correctly", func(t *testing.T) {
		svc := &mockAssetService{
			searchAssetsFunc: func(_ context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
				if cfg.Filters["type"][0] != "table" {
					t.Errorf("expected type filter 'table', got %v", cfg.Filters["type"])
				}
				if cfg.Filters["service"][0] != "bigquery" {
					t.Errorf("expected service filter 'bigquery', got %v", cfg.Filters["service"])
				}
				return []asset.SearchResult{}, nil
			},
		}
		s := newTestServer(svc)
		_, err := s.handleSearchAssets(ctx, callToolRequest(map[string]any{
			"text":     "test",
			"types":    "table",
			"services": "bigquery",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := &mockAssetService{
			searchAssetsFunc: func(_ context.Context, _ asset.SearchConfig) ([]asset.SearchResult, error) {
				return nil, fmt.Errorf("connection refused")
			},
		}
		s := newTestServer(svc)
		result, err := s.handleSearchAssets(ctx, callToolRequest(map[string]any{"text": "test"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleGetAsset(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when id is empty", func(t *testing.T) {
		s := newTestServer(&mockAssetService{})
		result, err := s.handleGetAsset(ctx, callToolRequest(map[string]any{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("returns asset details", func(t *testing.T) {
		svc := &mockAssetService{
			getAssetByIDFunc: func(_ context.Context, id string) (asset.Asset, error) {
				return asset.Asset{
					ID:          "123",
					URN:         "urn:bq:orders",
					Name:        "orders",
					Type:        asset.Type("table"),
					Service:     "bigquery",
					Description: "Main orders table",
					Owners:      []user.User{{Email: "alice@co.com"}},
				}, nil
			},
		}
		s := newTestServer(svc)
		result, err := s.handleGetAsset(ctx, callToolRequest(map[string]any{"id": "123"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := getTextContent(result)
		if text == "" {
			t.Error("expected non-empty text content")
		}
	})
}

func TestHandleGetLineage(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when urn is empty", func(t *testing.T) {
		s := newTestServer(&mockAssetService{})
		result, err := s.handleGetLineage(ctx, callToolRequest(map[string]any{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("returns lineage", func(t *testing.T) {
		svc := &mockAssetService{
			getLineageFunc: func(_ context.Context, urn string, q asset.LineageQuery) (asset.Lineage, error) {
				return asset.Lineage{
					Edges: []asset.LineageEdge{
						{Source: "urn:bq:raw_orders", Target: urn},
						{Source: urn, Target: "urn:bq:order_summary"},
					},
				}, nil
			},
		}
		s := newTestServer(svc)
		result, err := s.handleGetLineage(ctx, callToolRequest(map[string]any{"urn": "urn:bq:orders"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := getTextContent(result)
		if text == "" {
			t.Error("expected non-empty text content")
		}
	})

	t.Run("validates direction", func(t *testing.T) {
		s := newTestServer(&mockAssetService{})
		result, err := s.handleGetLineage(ctx, callToolRequest(map[string]any{
			"urn":       "urn:bq:orders",
			"direction": "invalid",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result for invalid direction")
		}
	})
}

func TestHandleListTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("returns types", func(t *testing.T) {
		svc := &mockAssetService{
			getTypesFunc: func(_ context.Context, _ asset.Filter) (map[asset.Type]int, error) {
				return map[asset.Type]int{
					"table": 42,
					"topic": 10,
				}, nil
			},
		}
		s := newTestServer(svc)
		result, err := s.handleListTypes(ctx, callToolRequest(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := getTextContent(result)
		if text == "" {
			t.Error("expected non-empty text content")
		}
	})
}

func TestHandleGetAllAssets(t *testing.T) {
	ctx := context.Background()

	t.Run("returns assets with pagination", func(t *testing.T) {
		svc := &mockAssetService{
			getAllAssetsFunc: func(_ context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error) {
				if !withTotal {
					t.Error("expected withTotal to be true")
				}
				return []asset.Asset{
					{ID: "1", URN: "urn:bq:orders", Name: "orders", Type: "table", Service: "bigquery"},
				}, 100, nil
			},
		}
		s := newTestServer(svc)
		result, err := s.handleGetAllAssets(ctx, callToolRequest(map[string]any{"size": 10}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := getTextContent(result)
		if text == "" {
			t.Error("expected non-empty text content")
		}
	})
}

// getTextContent extracts text from the first TextContent in a CallToolResult.
func getTextContent(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
