package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/namespace"
)

type DocumentRepository struct {
	client *Client
}

func NewDocumentRepository(client *Client) (*DocumentRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &DocumentRepository{client: client}, nil
}

func (r *DocumentRepository) Upsert(ctx context.Context, ns *namespace.Namespace, doc *document.Document) (string, error) {
	now := time.Now().UTC()
	doc.UpdatedAt = now

	var id string
	err := r.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx,
			`INSERT INTO documents (namespace_id, entity_urn, title, body, format, source, source_id, properties, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (namespace_id, entity_urn, source, source_id)
			 DO UPDATE SET title = EXCLUDED.title, body = EXCLUDED.body, format = EXCLUDED.format,
			              properties = EXCLUDED.properties, updated_at = EXCLUDED.updated_at
			 RETURNING id`,
			ns.ID, doc.EntityURN, doc.Title, doc.Body, doc.Format,
			doc.Source, nilIfEmpty(doc.SourceID), JSONMap(doc.Properties), now, now,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("upsert document: %w", err)
	}
	doc.CreatedAt = now
	return id, nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (document.Document, error) {
	var m documentModel
	err := r.client.GetContext(ctx, &m,
		`SELECT id, namespace_id, entity_urn, title, body, format,
				COALESCE(source, '') as source, COALESCE(source_id, '') as source_id,
				properties, created_at, updated_at
		 FROM documents WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return document.Document{}, err
		}
		return document.Document{}, fmt.Errorf("get document: %w", err)
	}
	return m.toDomain(), nil
}

func (r *DocumentRepository) GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]document.Document, error) {
	var models []documentModel
	err := r.client.SelectContext(ctx, &models,
		`SELECT id, namespace_id, entity_urn, title, body, format,
				COALESCE(source, '') as source, COALESCE(source_id, '') as source_id,
				properties, created_at, updated_at
		 FROM documents WHERE namespace_id = $1 AND entity_urn = $2
		 ORDER BY created_at`, ns.ID, entityURN)
	if err != nil {
		return nil, fmt.Errorf("get documents by entity: %w", err)
	}
	return toDocuments(models), nil
}

func (r *DocumentRepository) GetAll(ctx context.Context, ns *namespace.Namespace, filter document.Filter) ([]document.Document, error) {
	builder := sq.Select("id", "namespace_id", "entity_urn", "title", "body", "format",
		"COALESCE(source, '') as source", "COALESCE(source_id, '') as source_id",
		"properties", "created_at", "updated_at").
		From("documents").
		Where(sq.Eq{"namespace_id": ns.ID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	if filter.EntityURN != "" {
		builder = builder.Where(sq.Eq{"entity_urn": filter.EntityURN})
	}
	if filter.Source != "" {
		builder = builder.Where(sq.Eq{"source": filter.Source})
	}

	limit := filter.Size
	if limit <= 0 {
		limit = 50
	}
	builder = builder.Limit(uint64(limit)).Offset(uint64(filter.Offset))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var models []documentModel
	if err := r.client.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("get all documents: %w", err)
	}
	return toDocuments(models), nil
}

func (r *DocumentRepository) Delete(ctx context.Context, ns *namespace.Namespace, id string) error {
	res, err := r.client.ExecContext(ctx,
		`DELETE FROM documents WHERE namespace_id = $1 AND id = $2`, ns.ID, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *DocumentRepository) DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error {
	_, err := r.client.ExecContext(ctx,
		`DELETE FROM documents WHERE namespace_id = $1 AND entity_urn = $2`, ns.ID, entityURN)
	return err
}

type documentModel struct {
	ID          string    `db:"id"`
	NamespaceID string    `db:"namespace_id"`
	EntityURN   string    `db:"entity_urn"`
	Title       string    `db:"title"`
	Body        string    `db:"body"`
	Format      string    `db:"format"`
	Source      string    `db:"source"`
	SourceID    string    `db:"source_id"`
	Properties  JSONMap   `db:"properties"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (m documentModel) toDomain() document.Document {
	return document.Document{
		ID:          m.ID,
		NamespaceID: m.NamespaceID,
		EntityURN:   m.EntityURN,
		Title:       m.Title,
		Body:        m.Body,
		Format:      m.Format,
		Source:      m.Source,
		SourceID:    m.SourceID,
		Properties:  map[string]interface{}(m.Properties),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toDocuments(models []documentModel) []document.Document {
	docs := make([]document.Document, len(models))
	for i, m := range models {
		docs[i] = m.toDomain()
	}
	return docs
}
