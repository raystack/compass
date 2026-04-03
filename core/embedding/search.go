package embedding

import (
	"cmp"
	"context"
	"slices"

	"github.com/raystack/compass/core/entity"
)

// HybridSearch fuses keyword (Postgres) + semantic (pgvector) results using RRF.
type HybridSearch struct {
	search  entity.SearchRepository
	repo    Repository
	embedFn EmbeddingFunc
}

func NewHybridSearch(search entity.SearchRepository, repo Repository, embedFn EmbeddingFunc) *HybridSearch {
	return &HybridSearch{search: search, repo: repo, embedFn: embedFn}
}

func (h *HybridSearch) Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
	switch cfg.Mode {
	case entity.SearchModeSemantic:
		return h.semanticSearch(ctx, cfg)
	case entity.SearchModeHybrid:
		return h.hybridSearch(ctx, cfg)
	default:
		return h.search.Search(ctx, cfg)
	}
}

func (h *HybridSearch) semanticSearch(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
	if h.embedFn == nil || h.repo == nil {
		return h.search.Search(ctx, cfg)
	}

	vec, err := h.embedFn(ctx, cfg.Text)
	if err != nil {
		return nil, err
	}

	limit := cfg.MaxResults
	if limit <= 0 {
		limit = 10
	}

	embeddings, err := h.repo.Search(ctx, cfg.Namespace, vec, limit)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var results []entity.SearchResult
	for _, e := range embeddings {
		if seen[e.EntityURN] {
			continue
		}
		seen[e.EntityURN] = true
		results = append(results, entity.SearchResult{URN: e.EntityURN, Description: e.Content})
	}
	return results, nil
}

func (h *HybridSearch) hybridSearch(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
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
func reciprocalRankFusion(lists ...[]entity.SearchResult) []entity.SearchResult {
	const k = 60.0

	type scored struct {
		result entity.SearchResult
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

	results := make([]entity.SearchResult, len(all))
	for i, s := range all {
		results[i] = s.result
	}
	return results
}
