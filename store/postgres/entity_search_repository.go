package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

// EntitySearchRepository implements entity.SearchRepository using Postgres
// tsvector (full-text search) + pg_trgm (fuzzy matching). No ES dependency.
type EntitySearchRepository struct {
	client *Client
}

func NewEntitySearchRepository(client *Client) (*EntitySearchRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &EntitySearchRepository{client: client}, nil
}

// Search performs keyword search using tsvector ranking with pg_trgm fuzzy fallback.
// Ranking: ts_rank weights URN/name (A) higher than description (B) and source (C).
// If tsvector returns no results, falls back to pg_trgm similarity on name and URN.
func (r *EntitySearchRepository) Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error) {
	limit := cfg.MaxResults
	if limit <= 0 {
		limit = 10
	}

	nsID := ""
	if cfg.Namespace != nil {
		nsID = cfg.Namespace.ID.String()
	}

	// Primary: tsvector full-text search with ranking
	results, err := r.tsvectorSearch(ctx, nsID, cfg.Text, cfg.Filters, limit, cfg.Offset)
	if err != nil {
		return nil, err
	}

	// Fallback: if tsvector returned nothing, try pg_trgm fuzzy match
	if len(results) == 0 && cfg.Text != "" {
		results, err = r.trigramSearch(ctx, nsID, cfg.Text, cfg.Filters, limit, cfg.Offset)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// Suggest returns name completions using pg_trgm similarity.
func (r *EntitySearchRepository) Suggest(ctx context.Context, ns *namespace.Namespace, text string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `SELECT DISTINCT name FROM entities
		WHERE namespace_id = $1 AND valid_to IS NULL AND name % $2
		ORDER BY similarity(name, $2) DESC
		LIMIT $3`

	var names []string
	if err := r.client.SelectContext(ctx, &names, query, ns.ID, text, limit); err != nil {
		return nil, fmt.Errorf("suggest: %w", err)
	}
	return names, nil
}

func (r *EntitySearchRepository) tsvectorSearch(ctx context.Context, nsID, text string, filters map[string][]string, limit, offset int) ([]entity.SearchResult, error) {
	// Build the query with plainto_tsquery for robustness (handles unquoted input)
	query := `SELECT id, urn, type, name, COALESCE(source, '') as source,
			COALESCE(description, '') as description,
			ts_rank(search_vector, plainto_tsquery('english', $2)) as rank
		FROM entities
		WHERE namespace_id = $1 AND valid_to IS NULL
			AND search_vector @@ plainto_tsquery('english', $2)`

	args := []interface{}{nsID, text}
	argIdx := 3

	// Apply type filter
	if types, ok := filters["type"]; ok && len(types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argIdx)
		args = append(args, types)
		argIdx++
	}

	// Apply source filter
	if sources, ok := filters["source"]; ok && len(sources) > 0 {
		query += fmt.Sprintf(" AND source = ANY($%d)", argIdx)
		args = append(args, sources)
		argIdx++
	}

	query += " ORDER BY rank DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	return r.querySearchResults(ctx, query, args...)
}

func (r *EntitySearchRepository) trigramSearch(ctx context.Context, nsID, text string, filters map[string][]string, limit, offset int) ([]entity.SearchResult, error) {
	// pg_trgm similarity search: matches even with typos
	query := `SELECT id, urn, type, name, COALESCE(source, '') as source,
			COALESCE(description, '') as description,
			GREATEST(similarity(name, $2), similarity(urn, $2)) as rank
		FROM entities
		WHERE namespace_id = $1 AND valid_to IS NULL
			AND (name % $2 OR urn % $2)`

	args := []interface{}{nsID, text}
	argIdx := 3

	if types, ok := filters["type"]; ok && len(types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argIdx)
		args = append(args, types)
		argIdx++
	}

	if sources, ok := filters["source"]; ok && len(sources) > 0 {
		query += fmt.Sprintf(" AND source = ANY($%d)", argIdx)
		args = append(args, sources)
		argIdx++
	}

	query += " ORDER BY rank DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	return r.querySearchResults(ctx, query, args...)
}

func (r *EntitySearchRepository) querySearchResults(ctx context.Context, query string, args ...interface{}) ([]entity.SearchResult, error) {
	type row struct {
		ID          string  `db:"id"`
		URN         string  `db:"urn"`
		Type        string  `db:"type"`
		Name        string  `db:"name"`
		Source      string  `db:"source"`
		Description string  `db:"description"`
		Rank        float64 `db:"rank"`
	}

	var rows []row
	if err := r.client.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("search entities: %w", err)
	}

	results := make([]entity.SearchResult, len(rows))
	for i, r := range rows {
		results[i] = entity.SearchResult{
			ID:          r.ID,
			URN:         r.URN,
			Type:        r.Type,
			Name:        r.Name,
			Source:      r.Source,
			Description: r.Description,
			Rank:        r.Rank,
		}
	}
	return results, nil
}
