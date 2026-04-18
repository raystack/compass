package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

// --- Mock services ---

type mockEntityService struct {
	searchFn          func(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error)
	getContextFn      func(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error)
	getImpactFn       func(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error)
	assembleContextFn func(ctx context.Context, ns *namespace.Namespace, req entity.AssemblyRequest) (*entity.AssembledContext, error)
}

func (m *mockEntityService) Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
	return m.searchFn(ctx, cfg)
}

func (m *mockEntityService) GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error) {
	return m.getContextFn(ctx, ns, urn, depth)
}

func (m *mockEntityService) GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error) {
	return m.getImpactFn(ctx, ns, urn, depth)
}

func (m *mockEntityService) AssembleContext(ctx context.Context, ns *namespace.Namespace, req entity.AssemblyRequest) (*entity.AssembledContext, error) {
	return m.assembleContextFn(ctx, ns, req)
}

type mockDocumentService struct {
	getByEntityURNFn func(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]document.Document, error)
}

func (m *mockDocumentService) GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]document.Document, error) {
	return m.getByEntityURNFn(ctx, ns, entityURN)
}

// --- Helpers ---

func makeRequest(args map[string]any) gomcp.CallToolRequest {
	return gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Arguments: args,
		},
	}
}

func newTestServer(entitySvc EntityService, docSvc DocumentService) *Server {
	return &Server{
		entityService:   entitySvc,
		documentService: docSvc,
	}
}

func resultText(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}
	tc, ok := result.Content[0].(gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// --- Entity handler tests ---

func TestHandleSearchEntities(t *testing.T) {
	svc := &mockEntityService{
		searchFn: func(_ context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
			if cfg.Text != "orders" {
				t.Errorf("expected text 'orders', got %q", cfg.Text)
			}
			return []entity.SearchResult{
				{Name: "orders_table", Type: "table", Source: "bigquery", URN: "urn:bq:orders", Description: "Order events"},
				{Name: "orders_topic", Type: "topic", Source: "kafka", URN: "urn:kafka:orders", Description: ""},
			}, nil
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleSearchEntities(context.Background(), makeRequest(map[string]any{
		"text": "orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "Found 2 entities") {
		t.Errorf("expected 'Found 2 entities' in output, got: %s", text)
	}
	if !strings.Contains(text, "orders_table") {
		t.Errorf("expected 'orders_table' in output")
	}
	if !strings.Contains(text, "Order events") {
		t.Errorf("expected description in output")
	}
}

func TestHandleSearchEntities_MissingText(t *testing.T) {
	srv := newTestServer(&mockEntityService{}, nil)

	result, err := srv.handleSearchEntities(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "'text' parameter is required") {
		t.Errorf("expected text parameter error, got: %s", text)
	}
}

func TestHandleSearchEntities_WithFilters(t *testing.T) {
	var capturedCfg entity.SearchConfig
	svc := &mockEntityService{
		searchFn: func(_ context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
			capturedCfg = cfg
			return nil, nil
		},
	}
	srv := newTestServer(svc, nil)

	_, err := srv.handleSearchEntities(context.Background(), makeRequest(map[string]any{
		"text":   "foo",
		"types":  "table,topic",
		"source": "bigquery",
		"mode":   "semantic",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedCfg.Mode != entity.SearchModeSemantic {
		t.Errorf("expected mode semantic, got %q", capturedCfg.Mode)
	}
	types, ok := capturedCfg.Filters["type"]
	if !ok {
		t.Fatal("expected type filter")
	}
	if len(types) != 2 || types[0] != "table" || types[1] != "topic" {
		t.Errorf("unexpected type filter: %v", types)
	}
	sources, ok := capturedCfg.Filters["source"]
	if !ok {
		t.Fatal("expected source filter")
	}
	if len(sources) != 1 || sources[0] != "bigquery" {
		t.Errorf("unexpected source filter: %v", sources)
	}
}

func TestHandleGetContext(t *testing.T) {
	svc := &mockEntityService{
		getContextFn: func(_ context.Context, _ *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error) {
			if urn != "urn:bq:orders" {
				t.Errorf("expected urn 'urn:bq:orders', got %q", urn)
			}
			if depth != 2 {
				t.Errorf("expected default depth 2, got %d", depth)
			}
			return &entity.ContextGraph{
				Entity: entity.Entity{
					Name:        "orders_table",
					Type:        "table",
					URN:         "urn:bq:orders",
					Source:      "bigquery",
					Description: "All orders",
				},
				Edges: []entity.Edge{
					{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:dashboard", Type: "lineage"},
					{SourceURN: "urn:kafka:events", TargetURN: "urn:bq:orders", Type: "lineage"},
				},
				Related: []entity.Entity{
					{Name: "dashboard", Type: "dashboard", URN: "urn:bq:dashboard"},
					{Name: "events", Type: "topic", URN: "urn:kafka:events"},
				},
			}, nil
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleGetContext(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "orders_table") {
		t.Errorf("expected entity name in output")
	}
	if !strings.Contains(text, "Relationships") {
		t.Errorf("expected Relationships section")
	}
	if !strings.Contains(text, "Related Entities") {
		t.Errorf("expected Related Entities section")
	}
}

func TestHandleGetContext_MissingURN(t *testing.T) {
	srv := newTestServer(&mockEntityService{}, nil)

	result, err := srv.handleGetContext(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "'urn' parameter is required") {
		t.Errorf("expected urn parameter error, got: %s", text)
	}
}

func TestHandleImpact(t *testing.T) {
	svc := &mockEntityService{
		getImpactFn: func(_ context.Context, _ *namespace.Namespace, urn string, depth int) ([]entity.Edge, error) {
			if urn != "urn:bq:orders" {
				t.Errorf("expected urn 'urn:bq:orders', got %q", urn)
			}
			if depth != 3 {
				t.Errorf("expected default depth 3, got %d", depth)
			}
			return []entity.Edge{
				{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:dashboard", Type: "lineage"},
				{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:report", Type: "lineage"},
			}, nil
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleImpact(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "Impact Analysis") {
		t.Errorf("expected Impact Analysis header")
	}
	if !strings.Contains(text, "2 entities affected") {
		t.Errorf("expected '2 entities affected' in output, got: %s", text)
	}
}

func TestHandleImpact_MissingURN(t *testing.T) {
	srv := newTestServer(&mockEntityService{}, nil)

	result, err := srv.handleImpact(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "'urn' parameter is required") {
		t.Errorf("expected urn parameter error, got: %s", text)
	}
}

// --- Document handler tests ---

func TestHandleGetDocuments(t *testing.T) {
	docSvc := &mockDocumentService{
		getByEntityURNFn: func(_ context.Context, _ *namespace.Namespace, entityURN string) ([]document.Document, error) {
			if entityURN != "urn:bq:orders" {
				t.Errorf("expected urn 'urn:bq:orders', got %q", entityURN)
			}
			return []document.Document{
				{Title: "Runbook: Orders Table", Body: "Steps to debug orders table issues.", Source: "confluence", SourceID: "12345"},
				{Title: "Design Doc", Body: "Architecture of the orders pipeline.", Source: "github", SourceID: ""},
			}, nil
		},
	}
	srv := newTestServer(nil, docSvc)

	result, err := srv.handleGetDocuments(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "Found 2 documents") {
		t.Errorf("expected 'Found 2 documents' in output, got: %s", text)
	}
	if !strings.Contains(text, "Runbook: Orders Table") {
		t.Errorf("expected document title in output")
	}
	if !strings.Contains(text, "(ID: 12345)") {
		t.Errorf("expected source ID in output")
	}
}

func TestHandleGetDocuments_MissingURN(t *testing.T) {
	srv := newTestServer(nil, &mockDocumentService{})

	result, err := srv.handleGetDocuments(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "'urn' parameter is required") {
		t.Errorf("expected urn parameter error, got: %s", text)
	}
}

func TestHandleGetDocuments_NoDocuments(t *testing.T) {
	docSvc := &mockDocumentService{
		getByEntityURNFn: func(_ context.Context, _ *namespace.Namespace, _ string) ([]document.Document, error) {
			return nil, nil
		},
	}
	srv := newTestServer(nil, docSvc)

	result, err := srv.handleGetDocuments(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:empty",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "No documents found") {
		t.Errorf("expected 'No documents found' in output, got: %s", text)
	}
}

// --- Formatter tests ---

func TestFormatEntitySearchResults(t *testing.T) {
	results := []entity.SearchResult{
		{Name: "table_a", Type: "table", Source: "bigquery", URN: "urn:bq:a", Description: "First table"},
		{Name: "table_b", Type: "table", Source: "bigquery", URN: "urn:bq:b", Description: ""},
	}

	text := formatEntitySearchResults(results)

	if !strings.Contains(text, "Found 2 entities") {
		t.Errorf("expected header, got: %s", text)
	}
	if !strings.Contains(text, "**table_a** (table)") {
		t.Errorf("expected formatted entity name")
	}
	if !strings.Contains(text, "First table") {
		t.Errorf("expected description")
	}
	// table_b has no description, should not have an extra line
	if strings.Contains(text, "  \n") {
		t.Errorf("expected no blank description line for table_b")
	}
}

func TestFormatEntitySearchResults_Empty(t *testing.T) {
	text := formatEntitySearchResults(nil)
	if text != "No entities found." {
		t.Errorf("expected 'No entities found.', got: %q", text)
	}
}

func TestFormatContextGraph(t *testing.T) {
	cg := &entity.ContextGraph{
		Entity: entity.Entity{
			Name:        "orders",
			Type:        "table",
			URN:         "urn:bq:orders",
			Source:      "bigquery",
			Description: "All order events",
		},
		Edges: []entity.Edge{
			{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:dashboard", Type: "lineage"},
			{SourceURN: "urn:kafka:events", TargetURN: "urn:bq:orders", Type: "lineage"},
		},
		Related: []entity.Entity{
			{Name: "dashboard", Type: "dashboard", URN: "urn:bq:dashboard"},
		},
	}

	text := formatContextGraph(cg)

	if !strings.Contains(text, "## orders (table)") {
		t.Errorf("expected entity header")
	}
	if !strings.Contains(text, "URN: urn:bq:orders") {
		t.Errorf("expected URN line")
	}
	if !strings.Contains(text, "All order events") {
		t.Errorf("expected description")
	}
	// outgoing edge
	if !strings.Contains(text, "[lineage]") {
		t.Errorf("expected edge type in output")
	}
	if !strings.Contains(text, "urn:bq:dashboard") {
		t.Errorf("expected target URN in output")
	}
	// incoming edge
	if !strings.Contains(text, "urn:kafka:events") {
		t.Errorf("expected source URN in output")
	}
	// related
	if !strings.Contains(text, "Related Entities") {
		t.Errorf("expected Related Entities section")
	}
	if !strings.Contains(text, "**dashboard** (dashboard)") {
		t.Errorf("expected related entity")
	}
}

func TestFormatImpactAnalysis(t *testing.T) {
	edges := []entity.Edge{
		{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:dashboard", Type: "lineage"},
		{SourceURN: "urn:bq:dashboard", TargetURN: "urn:bq:report", Type: "lineage"},
	}

	text := formatImpactAnalysis("urn:bq:orders", edges)

	if !strings.Contains(text, "Impact Analysis for urn:bq:orders") {
		t.Errorf("expected impact analysis header")
	}
	if !strings.Contains(text, "2 entities affected") {
		t.Errorf("expected '2 entities affected', got: %s", text)
	}
	if !strings.Contains(text, "2 edges") {
		t.Errorf("expected '2 edges' in output")
	}
}

func TestFormatImpactAnalysis_NoEdges(t *testing.T) {
	text := formatImpactAnalysis("urn:bq:orders", nil)
	if !strings.Contains(text, "No downstream dependencies") {
		t.Errorf("expected 'No downstream dependencies', got: %q", text)
	}
}

// --- Error propagation tests ---

func TestHandleSearchEntities_ServiceError(t *testing.T) {
	svc := &mockEntityService{
		searchFn: func(_ context.Context, _ entity.SearchConfig) ([]entity.SearchResult, error) {
			return nil, errors.New("connection refused")
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleSearchEntities(context.Background(), makeRequest(map[string]any{
		"text": "orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "search failed") {
		t.Errorf("expected 'search failed' in error, got: %s", text)
	}
}

func TestHandleGetContext_ServiceError(t *testing.T) {
	svc := &mockEntityService{
		getContextFn: func(_ context.Context, _ *namespace.Namespace, _ string, _ int) (*entity.ContextGraph, error) {
			return nil, errors.New("not found")
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleGetContext(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:missing",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "get context failed") {
		t.Errorf("expected 'get context failed' in error, got: %s", text)
	}
}

func TestHandleImpact_ServiceError(t *testing.T) {
	svc := &mockEntityService{
		getImpactFn: func(_ context.Context, _ *namespace.Namespace, _ string, _ int) ([]entity.Edge, error) {
			return nil, errors.New("timeout")
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleImpact(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "impact analysis failed") {
		t.Errorf("expected 'impact analysis failed' in error, got: %s", text)
	}
}

func TestHandleGetDocuments_ServiceError(t *testing.T) {
	docSvc := &mockDocumentService{
		getByEntityURNFn: func(_ context.Context, _ *namespace.Namespace, _ string) ([]document.Document, error) {
			return nil, errors.New("db error")
		},
	}
	srv := newTestServer(nil, docSvc)

	result, err := srv.handleGetDocuments(context.Background(), makeRequest(map[string]any{
		"urn": "urn:bq:orders",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "get documents failed") {
		t.Errorf("expected 'get documents failed' in error, got: %s", text)
	}
}

// --- Assemble context handler tests ---

func TestHandleAssembleContext(t *testing.T) {
	svc := &mockEntityService{
		assembleContextFn: func(_ context.Context, _ *namespace.Namespace, req entity.AssemblyRequest) (*entity.AssembledContext, error) {
			if req.Query != "debug orders pipeline" {
				t.Errorf("expected query 'debug orders pipeline', got %q", req.Query)
			}
			if req.Intent != entity.IntentDebug {
				t.Errorf("expected intent debug, got %q", req.Intent)
			}
			return &entity.AssembledContext{
				Query:       req.Query,
				Intent:      req.Intent,
				TokenBudget: 4000,
				TokensUsed:  500,
				Seeds: []entity.Entity{
					{Name: "orders", Type: "table", URN: "urn:bq:orders", Source: "bigquery", Description: "Order events"},
				},
				Entities: []entity.ScoredEntity{
					{Entity: entity.Entity{Name: "orders", Type: "table", URN: "urn:bq:orders"}, Score: 1.0, Distance: 0},
					{Entity: entity.Entity{Name: "dashboard", Type: "dashboard", URN: "urn:bq:dashboard"}, Score: 0.5, Distance: 1},
				},
				Edges: []entity.Edge{
					{SourceURN: "urn:bq:orders", TargetURN: "urn:bq:dashboard", Type: "lineage"},
				},
				Documents: []entity.FetchedDocument{
					{Title: "Runbook", Body: "Steps to debug.", EntityURN: "urn:bq:orders"},
				},
				Stats: entity.AssemblyStats{
					EntitiesConsidered: 5,
					EntitiesIncluded:   2,
					DocumentsFetched:   1,
					GraphDepth:         2,
				},
			}, nil
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleAssembleContext(context.Background(), makeRequest(map[string]any{
		"query":  "debug orders pipeline",
		"intent": "debug",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, result))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "Context: debug orders pipeline") {
		t.Errorf("expected context header in output")
	}
	if !strings.Contains(text, "Seed Entities") {
		t.Errorf("expected Seed Entities section")
	}
	if !strings.Contains(text, "Related Entities") {
		t.Errorf("expected Related Entities section")
	}
	if !strings.Contains(text, "Relationships") {
		t.Errorf("expected Relationships section")
	}
	if !strings.Contains(text, "Documents") {
		t.Errorf("expected Documents section")
	}
	if !strings.Contains(text, "Provenance") {
		t.Errorf("expected Provenance line")
	}
}

func TestHandleAssembleContext_MissingQuery(t *testing.T) {
	srv := newTestServer(&mockEntityService{}, nil)

	result, err := srv.handleAssembleContext(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "'query' parameter is required") {
		t.Errorf("expected query parameter error, got: %s", text)
	}
}

func TestHandleAssembleContext_ServiceError(t *testing.T) {
	svc := &mockEntityService{
		assembleContextFn: func(_ context.Context, _ *namespace.Namespace, _ entity.AssemblyRequest) (*entity.AssembledContext, error) {
			return nil, errors.New("db failure")
		},
	}
	srv := newTestServer(svc, nil)

	result, err := srv.handleAssembleContext(context.Background(), makeRequest(map[string]any{
		"query": "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "assemble context failed") {
		t.Errorf("expected 'assemble context failed' in error, got: %s", text)
	}
}
