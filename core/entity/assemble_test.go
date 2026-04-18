package entity

import (
	"context"
	"testing"

	"github.com/raystack/compass/core/namespace"
)

// mockDocFetcher implements DocumentFetcher for tests.
type mockDocFetcher struct {
	docs map[string][]FetchedDocument
}

func (m *mockDocFetcher) GetDocumentsForEntity(_ context.Context, _ *namespace.Namespace, entityURN string) ([]FetchedDocument, error) {
	return m.docs[entityURN], nil
}

func TestAssembleContext_WithSeedURNs(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "table_a", Source: "bigquery", Description: "Main table"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeDashboard, Name: "dash_b", Source: "metabase"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeJob, Name: "job_c", Source: "airflow"})
	edges.edges = []Edge{
		{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
		{SourceURN: "urn:a", TargetURN: "urn:c", Type: "lineage"},
	}

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:       "understand table_a",
		SeedURNs:    []string{"urn:a"},
		Intent:      IntentGeneral,
		TokenBudget: 4000,
		Depth:       2,
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if len(result.Seeds) != 1 {
		t.Errorf("expected 1 seed, got %d", len(result.Seeds))
	}
	if result.Seeds[0].URN != "urn:a" {
		t.Errorf("expected seed urn:a, got %s", result.Seeds[0].URN)
	}
	if len(result.Entities) < 2 {
		t.Errorf("expected at least 2 entities (seed + related), got %d", len(result.Entities))
	}
	if len(result.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(result.Edges))
	}
	if result.TokensUsed <= 0 {
		t.Errorf("expected positive tokens used, got %d", result.TokensUsed)
	}
	if result.Intent != IntentGeneral {
		t.Errorf("expected intent general, got %s", result.Intent)
	}
}

func TestAssembleContext_SearchSeeds(t *testing.T) {
	repo := newMockRepo()
	search := &mockSearchRepo{
		results: []SearchResult{
			{URN: "urn:table:orders", Name: "orders", Type: "table", Source: "bigquery", Description: "Orders"},
		},
	}
	svc := NewService(repo, nil, search)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query: "orders table",
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if len(result.Seeds) != 1 {
		t.Errorf("expected 1 seed from search, got %d", len(result.Seeds))
	}
	if result.Seeds[0].URN != "urn:table:orders" {
		t.Errorf("expected seed urn:table:orders, got %s", result.Seeds[0].URN)
	}
}

func TestAssembleContext_NoSeedsFound(t *testing.T) {
	repo := newMockRepo()
	search := &mockSearchRepo{results: nil}
	svc := NewService(repo, nil, search)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query: "nonexistent",
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if len(result.Seeds) != 0 {
		t.Errorf("expected 0 seeds, got %d", len(result.Seeds))
	}
	if len(result.Entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(result.Entities))
	}
	if result.Query != "nonexistent" {
		t.Errorf("expected query preserved, got %q", result.Query)
	}
}

func TestAssembleContext_TokenBudgetExceeded(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a", Description: "short"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b", Description: "related entity"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeTable, Name: "c", Description: "another entity"})
	edges.edges = []Edge{
		{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
		{SourceURN: "urn:a", TargetURN: "urn:c", Type: "lineage"},
	}

	// Very small budget - should truncate
	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:       "test",
		SeedURNs:    []string{"urn:a"},
		TokenBudget: 10, // tiny budget
		Depth:       2,
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if !result.Truncated {
		t.Error("expected truncated=true with tiny budget")
	}
}

func TestAssembleContext_WithoutDocumentFetcher(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:    "test",
		SeedURNs: []string{"urn:a"},
	})
	if err != nil {
		t.Fatalf("AssembleContext without docs should not error: %v", err)
	}

	if len(result.Documents) != 0 {
		t.Errorf("expected 0 documents without fetcher, got %d", len(result.Documents))
	}
	if len(result.Seeds) != 1 {
		t.Errorf("expected 1 seed, got %d", len(result.Seeds))
	}
}

func TestAssembleContext_WithDocuments(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	svc.WithDocumentFetcher(&mockDocFetcher{
		docs: map[string][]FetchedDocument{
			"urn:a": {
				{Title: "Runbook", Body: "How to fix table A.", Source: "confluence", EntityURN: "urn:a"},
			},
		},
	})

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:    "debug table a",
		SeedURNs: []string{"urn:a"},
		Intent:   IntentDebug,
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if len(result.Documents) != 1 {
		t.Errorf("expected 1 document, got %d", len(result.Documents))
	}
	if result.Documents[0].Title != "Runbook" {
		t.Errorf("expected document title 'Runbook', got %q", result.Documents[0].Title)
	}
}

func TestAssembleContext_IntentAffectsScoring(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:seed", Type: TypeTable, Name: "seed"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:dashboard", Type: TypeDashboard, Name: "dash"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:job", Type: TypeJob, Name: "job"})
	edges.edges = []Edge{
		{SourceURN: "urn:seed", TargetURN: "urn:dashboard", Type: "lineage"},
		{SourceURN: "urn:seed", TargetURN: "urn:job", Type: "lineage"},
	}

	// With analyze intent, dashboard should rank higher than job
	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:    "analyze",
		SeedURNs: []string{"urn:seed"},
		Intent:   IntentAnalyze,
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	// Find the dashboard and job scores
	var dashScore, jobScore float64
	for _, e := range result.Entities {
		if e.Entity.URN == "urn:dashboard" {
			dashScore = e.Score
		}
		if e.Entity.URN == "urn:job" {
			jobScore = e.Score
		}
	}

	if dashScore <= jobScore {
		t.Errorf("expected dashboard score (%.2f) > job score (%.2f) for analyze intent", dashScore, jobScore)
	}
}

func TestAssembleContext_DefaultsApplied(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})

	result, err := svc.AssembleContext(ctx, ns, AssemblyRequest{
		Query:    "test defaults",
		SeedURNs: []string{"urn:a"},
	})
	if err != nil {
		t.Fatalf("AssembleContext failed: %v", err)
	}

	if result.TokenBudget != 4000 {
		t.Errorf("expected default token budget 4000, got %d", result.TokenBudget)
	}
	if result.Intent != IntentGeneral {
		t.Errorf("expected default intent general, got %s", result.Intent)
	}
}

// --- Helper function tests ---

func TestEstimateTokens(t *testing.T) {
	if got := estimateTokens(""); got != 0 {
		t.Errorf("expected 0 for empty string, got %d", got)
	}
	if got := estimateTokens("abcd"); got != 1 {
		t.Errorf("expected 1 for 4-char string, got %d", got)
	}
	if got := estimateTokens("hello world"); got != 3 {
		t.Errorf("expected 3 for 11-char string, got %d", got)
	}
}

func TestValidateAssemblyRequest(t *testing.T) {
	req := validateAssemblyRequest(AssemblyRequest{})
	if req.TokenBudget != 4000 {
		t.Errorf("expected default budget 4000, got %d", req.TokenBudget)
	}
	if req.Depth != 2 {
		t.Errorf("expected default depth 2, got %d", req.Depth)
	}
	if req.Intent != IntentGeneral {
		t.Errorf("expected default intent general, got %s", req.Intent)
	}

	// Depth capped at 5
	req = validateAssemblyRequest(AssemblyRequest{Depth: 10})
	if req.Depth != 5 {
		t.Errorf("expected capped depth 5, got %d", req.Depth)
	}
}

func TestDeduplicateEdges(t *testing.T) {
	edges := []Edge{
		{SourceURN: "a", TargetURN: "b", Type: "lineage"},
		{SourceURN: "a", TargetURN: "b", Type: "lineage"},
		{SourceURN: "a", TargetURN: "c", Type: "lineage"},
	}
	deduped := deduplicateEdges(edges)
	if len(deduped) != 2 {
		t.Errorf("expected 2 unique edges, got %d", len(deduped))
	}
}

func TestIntentWeight(t *testing.T) {
	// Debug: job gets 1.5x
	if w := intentWeight(IntentDebug, TypeJob, 1); w != 1.5 {
		t.Errorf("expected 1.5 for debug+job, got %.2f", w)
	}
	// Analyze: dashboard gets 1.5x
	if w := intentWeight(IntentAnalyze, TypeDashboard, 1); w != 1.5 {
		t.Errorf("expected 1.5 for analyze+dashboard, got %.2f", w)
	}
	// General: all equal at 1.0
	if w := intentWeight(IntentGeneral, TypeTable, 1); w != 1.0 {
		t.Errorf("expected 1.0 for general+table, got %.2f", w)
	}
}
