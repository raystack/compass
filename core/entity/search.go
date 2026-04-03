package entity

import (
	"context"

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
