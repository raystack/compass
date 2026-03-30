package entity

import (
	"cmp"
	"context"
	"slices"

	"github.com/raystack/compass/core/namespace"
)

// SearchMode specifies which search backend(s) to use.
type SearchMode string

const (
	SearchModeKeyword  SearchMode = "keyword"  // Postgres tsvector + pg_trgm
	SearchModeSemantic SearchMode = "semantic"  // pgvector cosine similarity
	SearchModeHybrid   SearchMode = "hybrid"    // both, fused with RRF
)

// SearchConfig for entity search.
type SearchConfig struct {
	Text       string
	Filters    map[string][]string
	MaxResults int
	Offset     int
	Mode       SearchMode
	Namespace  *namespace.Namespace
}

// SearchResult represents a single search hit.
type SearchResult struct {
	ID          string  `json:"id"`
	URN         string  `json:"urn"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Source      string  `json:"source"`
	Description string  `json:"description"`
	Rank        float64 `json:"rank,omitempty"`
}

// SearchRepository defines search operations for entities.
// All implementations are Postgres-native (no ES dependency).
type SearchRepository interface {
	// Search performs keyword search using tsvector ranking + pg_trgm fuzzy matching.
	Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error)
	// Suggest returns name completions using pg_trgm similarity.
	Suggest(ctx context.Context, ns *namespace.Namespace, text string, limit int) ([]string, error)
}

// EmbeddingFunc generates vector embeddings from text.
type EmbeddingFunc func(ctx context.Context, text string) ([]float32, error)

// HybridSearch fuses keyword (Postgres) + semantic (pgvector) results using RRF.
type HybridSearch struct {
	search  SearchRepository
	chunks  ChunkRepository
	embedFn EmbeddingFunc
}

func NewHybridSearch(search SearchRepository, chunks ChunkRepository, embedFn EmbeddingFunc) *HybridSearch {
	return &HybridSearch{search: search, chunks: chunks, embedFn: embedFn}
}

func (h *HybridSearch) Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	switch cfg.Mode {
	case SearchModeSemantic:
		return h.semanticSearch(ctx, cfg)
	case SearchModeHybrid:
		return h.hybridSearch(ctx, cfg)
	default:
		return h.search.Search(ctx, cfg)
	}
}

func (h *HybridSearch) semanticSearch(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	if h.embedFn == nil || h.chunks == nil {
		return h.search.Search(ctx, cfg)
	}

	embedding, err := h.embedFn(ctx, cfg.Text)
	if err != nil {
		return nil, err
	}

	limit := cfg.MaxResults
	if limit <= 0 {
		limit = 10
	}

	chunks, err := h.chunks.Search(ctx, cfg.Namespace, embedding, limit)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var results []SearchResult
	for _, c := range chunks {
		if seen[c.EntityURN] {
			continue
		}
		seen[c.EntityURN] = true
		results = append(results, SearchResult{URN: c.EntityURN, Description: c.Content})
	}
	return results, nil
}

func (h *HybridSearch) hybridSearch(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	keywordResults, err := h.search.Search(ctx, cfg)
	if err != nil {
		return nil, err
	}

	semanticResults, err := h.semanticSearch(ctx, cfg)
	if err != nil {
		return keywordResults, nil // degrade gracefully
	}

	fused := reciprocalRankFusion(keywordResults, semanticResults)

	limit := cfg.MaxResults
	if limit <= 0 {
		limit = 10
	}
	if len(fused) > limit {
		fused = fused[:limit]
	}
	return fused, nil
}

// reciprocalRankFusion merges ranked lists. RRF score = Σ(1 / (k + rank)).
func reciprocalRankFusion(lists ...[]SearchResult) []SearchResult {
	const k = 60.0

	type scored struct {
		result SearchResult
		score  float64
	}

	scores := make(map[string]*scored)
	for _, list := range lists {
		for rank, r := range list {
			key := r.URN
			if key == "" {
				key = r.ID
			}
			s, ok := scores[key]
			if !ok {
				s = &scored{result: r}
				scores[key] = s
			}
			s.score += 1.0 / (k + float64(rank+1))
		}
	}

	all := make([]scored, 0, len(scores))
	for _, s := range scores {
		all = append(all, *s)
	}
	slices.SortFunc(all, func(a, b scored) int { return cmp.Compare(b.score, a.score) })

	results := make([]SearchResult, len(all))
	for i, s := range all {
		results[i] = s.result
	}
	return results
}
