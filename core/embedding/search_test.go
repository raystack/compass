package embedding

import (
	"context"
	"testing"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

func TestReciprocalRankFusion(t *testing.T) {
	list1 := []entity.SearchResult{
		{URN: "a", Name: "Alpha"},
		{URN: "b", Name: "Beta"},
		{URN: "c", Name: "Charlie"},
	}
	list2 := []entity.SearchResult{
		{URN: "b", Name: "Beta"},
		{URN: "d", Name: "Delta"},
		{URN: "a", Name: "Alpha"},
	}

	fused := reciprocalRankFusion(list1, list2)

	if len(fused) != 4 {
		t.Fatalf("expected 4 results, got %d", len(fused))
	}

	// "a" and "b" appear in both lists, should be ranked higher
	topTwo := map[string]bool{fused[0].URN: true, fused[1].URN: true}
	if !topTwo["a"] || !topTwo["b"] {
		t.Errorf("expected a and b in top 2, got %s and %s", fused[0].URN, fused[1].URN)
	}
}

func TestReciprocalRankFusion_EmptyLists(t *testing.T) {
	fused := reciprocalRankFusion(nil, nil)
	if len(fused) != 0 {
		t.Fatalf("expected 0 results, got %d", len(fused))
	}
}

func TestReciprocalRankFusion_SingleList(t *testing.T) {
	list := []entity.SearchResult{
		{URN: "a"},
		{URN: "b"},
	}
	fused := reciprocalRankFusion(list)
	if len(fused) != 2 {
		t.Fatalf("expected 2 results, got %d", len(fused))
	}
	if fused[0].URN != "a" {
		t.Errorf("expected first result to be 'a', got %q", fused[0].URN)
	}
}

// mockSearchRepo is a simple in-memory search repository for testing HybridSearch.
type mockSearchRepo struct {
	results     []entity.SearchResult
	suggestions []string
}

func (m *mockSearchRepo) Search(_ context.Context, _ entity.SearchConfig) ([]entity.SearchResult, error) {
	return m.results, nil
}

func (m *mockSearchRepo) Suggest(_ context.Context, _ *namespace.Namespace, _ string, _ int) ([]string, error) {
	return m.suggestions, nil
}

// mockEmbeddingRepo is a simple in-memory embedding repository for testing.
type mockEmbeddingRepo struct {
	embeddings []Embedding
}

func (m *mockEmbeddingRepo) UpsertBatch(_ context.Context, _ *namespace.Namespace, _ []Embedding) error {
	return nil
}

func (m *mockEmbeddingRepo) DeleteByEntityURN(_ context.Context, _ *namespace.Namespace, _ string) error {
	return nil
}

func (m *mockEmbeddingRepo) DeleteByContentID(_ context.Context, _ *namespace.Namespace, _ string) error {
	return nil
}

func (m *mockEmbeddingRepo) Search(_ context.Context, _ *namespace.Namespace, _ []float32, _ int) ([]Embedding, error) {
	return m.embeddings, nil
}

func TestHybridSearch_KeywordMode(t *testing.T) {
	search := &mockSearchRepo{
		results: []entity.SearchResult{
			{URN: "urn:table:orders", Name: "orders"},
			{URN: "urn:table:users", Name: "users"},
		},
	}
	hs := NewHybridSearch(search, nil, nil)
	ctx := context.Background()

	results, err := hs.Search(ctx, entity.SearchConfig{
		Text: "orders",
		Mode: entity.SearchModeKeyword,
	})
	if err != nil {
		t.Fatalf("keyword search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].URN != "urn:table:orders" {
		t.Errorf("expected first result URN 'urn:table:orders', got %q", results[0].URN)
	}
}

func TestHybridSearch_SemanticMode(t *testing.T) {
	search := &mockSearchRepo{}
	repo := &mockEmbeddingRepo{
		embeddings: []Embedding{
			{EntityURN: "urn:table:orders", Content: "orders table"},
			{EntityURN: "urn:table:payments", Content: "payments table"},
		},
	}
	embedFn := func(_ context.Context, text string) ([]float32, error) {
		return []float32{0.1, 0.2, 0.3}, nil
	}

	hs := NewHybridSearch(search, repo, embedFn)
	ctx := context.Background()

	results, err := hs.Search(ctx, entity.SearchConfig{
		Text:       "orders",
		Mode:       entity.SearchModeSemantic,
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("semantic search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].URN != "urn:table:orders" {
		t.Errorf("expected first result URN 'urn:table:orders', got %q", results[0].URN)
	}
}

func TestHybridSearch_HybridMode(t *testing.T) {
	search := &mockSearchRepo{
		results: []entity.SearchResult{
			{URN: "urn:table:orders", Name: "orders"},
			{URN: "urn:table:users", Name: "users"},
		},
	}
	repo := &mockEmbeddingRepo{
		embeddings: []Embedding{
			{EntityURN: "urn:table:orders", Content: "orders table"},
			{EntityURN: "urn:table:payments", Content: "payments table"},
		},
	}
	embedFn := func(_ context.Context, text string) ([]float32, error) {
		return []float32{0.1, 0.2, 0.3}, nil
	}

	hs := NewHybridSearch(search, repo, embedFn)
	ctx := context.Background()

	results, err := hs.Search(ctx, entity.SearchConfig{
		Text:       "orders",
		Mode:       entity.SearchModeHybrid,
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("hybrid search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}

	// "orders" appears in both keyword and semantic results, should be ranked first
	if results[0].URN != "urn:table:orders" {
		t.Errorf("expected 'urn:table:orders' to be top result (appears in both lists), got %q", results[0].URN)
	}

	// Should have 3 unique URNs total: orders, users, payments
	if len(results) != 3 {
		t.Errorf("expected 3 fused results, got %d", len(results))
	}
}
