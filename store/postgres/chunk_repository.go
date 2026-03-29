package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

type ChunkRepository struct {
	client *Client
}

func NewChunkRepository(client *Client) (*ChunkRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &ChunkRepository{client: client}, nil
}

func (r *ChunkRepository) UpsertBatch(ctx context.Context, ns *namespace.Namespace, chunks []entity.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Delete old chunks for affected entities, then insert fresh
	urns := uniqueURNs(chunks)
	for _, urn := range urns {
		if err := r.DeleteByEntityURN(ctx, ns, urn); err != nil {
			return fmt.Errorf("clear old chunks for %s: %w", urn, err)
		}
	}

	builder := sq.Insert("chunks").
		Columns("namespace_id", "entity_urn", "content", "context", "embedding", "position", "heading", "token_count").
		PlaceholderFormat(sq.Dollar)

	for _, c := range chunks {
		builder = builder.Values(ns.ID, c.EntityURN, c.Content, c.Context,
			vectorString(c.Embedding), c.Position, c.Heading, c.TokenCount)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build insert chunks: %w", err)
	}
	_, err = r.client.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert chunks: %w", err)
	}
	return nil
}

func (r *ChunkRepository) DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error {
	_, err := r.client.ExecContext(ctx,
		`DELETE FROM chunks WHERE namespace_id = $1 AND entity_urn = $2`, ns.ID, entityURN)
	return err
}

func (r *ChunkRepository) Search(ctx context.Context, ns *namespace.Namespace, embedding []float32, limit int) ([]entity.Chunk, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, namespace_id, entity_urn, content, context, position, heading, token_count, created_at
		FROM chunks WHERE namespace_id = $1
		ORDER BY embedding <=> $2::vector
		LIMIT $3`

	var models []chunkModel
	if err := r.client.SelectContext(ctx, &models, query, ns.ID, vectorString(embedding), limit); err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	result := make([]entity.Chunk, len(models))
	for i, m := range models {
		result[i] = entity.Chunk{
			ID:         m.ID,
			EntityURN:  m.EntityURN,
			Content:    m.Content,
			Context:    m.Context,
			Position:   m.Position,
			Heading:    m.Heading,
			TokenCount: m.TokenCount,
			CreatedAt:  m.CreatedAt,
		}
	}
	return result, nil
}

type chunkModel struct {
	ID         string    `db:"id"`
	EntityURN  string    `db:"entity_urn"`
	Content    string    `db:"content"`
	Context    string    `db:"context"`
	Position   int       `db:"position"`
	Heading    string    `db:"heading"`
	TokenCount int       `db:"token_count"`
	CreatedAt  time.Time `db:"created_at"`
}

func vectorString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func uniqueURNs(chunks []entity.Chunk) []string {
	seen := make(map[string]bool)
	var urns []string
	for _, c := range chunks {
		if !seen[c.EntityURN] {
			seen[c.EntityURN] = true
			urns = append(urns, c.EntityURN)
		}
	}
	return urns
}
