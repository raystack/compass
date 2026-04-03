package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/raystack/compass/core/embedding"
	"github.com/raystack/compass/core/namespace"
)

type EmbeddingRepository struct {
	client *Client
}

func NewEmbeddingRepository(client *Client) (*EmbeddingRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &EmbeddingRepository{client: client}, nil
}

func (r *EmbeddingRepository) UpsertBatch(ctx context.Context, ns *namespace.Namespace, embeddings []embedding.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	// Delete old embeddings for affected entities by content type, then insert fresh.
	// This avoids entity embeddings being overwritten when document embeddings are inserted.
	contentType := embeddings[0].ContentType
	urns := uniqueEmbeddingURNs(embeddings)
	for _, urn := range urns {
		_, err := r.client.ExecContext(ctx,
			`DELETE FROM embeddings WHERE namespace_id = $1 AND entity_urn = $2 AND content_type = $3`,
			ns.ID, urn, contentType)
		if err != nil {
			return fmt.Errorf("clear old embeddings for %s: %w", urn, err)
		}
	}

	builder := sq.Insert("embeddings").
		Columns("namespace_id", "entity_urn", "content_id", "content_type", "content", "context", "embedding", "position", "heading", "token_count").
		PlaceholderFormat(sq.Dollar)

	for _, e := range embeddings {
		builder = builder.Values(ns.ID, e.EntityURN, nilIfEmpty(e.ContentID), e.ContentType,
			e.Content, e.Context, vectorString(e.Vector), e.Position, e.Heading, e.TokenCount)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build insert embeddings: %w", err)
	}
	_, err = r.client.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert embeddings: %w", err)
	}
	return nil
}

func (r *EmbeddingRepository) DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error {
	_, err := r.client.ExecContext(ctx,
		`DELETE FROM embeddings WHERE namespace_id = $1 AND entity_urn = $2`, ns.ID, entityURN)
	return err
}

func (r *EmbeddingRepository) DeleteByContentID(ctx context.Context, ns *namespace.Namespace, contentID string) error {
	_, err := r.client.ExecContext(ctx,
		`DELETE FROM embeddings WHERE namespace_id = $1 AND content_id = $2`, ns.ID, contentID)
	return err
}

func (r *EmbeddingRepository) Search(ctx context.Context, ns *namespace.Namespace, vector []float32, limit int) ([]embedding.Embedding, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, entity_urn, COALESCE(content_id::text, '') as content_id,
			COALESCE(content_type, 'entity') as content_type, content, COALESCE(context, '') as context,
			position, COALESCE(heading, '') as heading, COALESCE(token_count, 0) as token_count, created_at
		FROM embeddings WHERE namespace_id = $1
		ORDER BY embedding <=> $2::vector
		LIMIT $3`

	var models []embeddingModel
	if err := r.client.SelectContext(ctx, &models, query, ns.ID, vectorString(vector), limit); err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	result := make([]embedding.Embedding, len(models))
	for i, m := range models {
		result[i] = embedding.Embedding{
			ID:          m.ID,
			EntityURN:   m.EntityURN,
			ContentID:   m.ContentID,
			ContentType: m.ContentType,
			Content:     m.Content,
			Context:     m.Context,
			Position:    m.Position,
			Heading:     m.Heading,
			TokenCount:  m.TokenCount,
			CreatedAt:   m.CreatedAt,
		}
	}
	return result, nil
}

type embeddingModel struct {
	ID          string    `db:"id"`
	EntityURN   string    `db:"entity_urn"`
	ContentID   string    `db:"content_id"`
	ContentType string    `db:"content_type"`
	Content     string    `db:"content"`
	Context     string    `db:"context"`
	Position    int       `db:"position"`
	Heading     string    `db:"heading"`
	TokenCount  int       `db:"token_count"`
	CreatedAt   time.Time `db:"created_at"`
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

func uniqueEmbeddingURNs(embeddings []embedding.Embedding) []string {
	seen := make(map[string]bool)
	var urns []string
	for _, e := range embeddings {
		if !seen[e.EntityURN] {
			seen[e.EntityURN] = true
			urns = append(urns, e.EntityURN)
		}
	}
	return urns
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
