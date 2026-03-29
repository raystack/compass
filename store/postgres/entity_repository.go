package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

type EntityRepository struct {
	client *Client
}

func NewEntityRepository(client *Client) (*EntityRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &EntityRepository{client: client}, nil
}

var entityColumns = `id, namespace_id, urn, type, name, description, properties, source,
	valid_from, valid_to, created_at, updated_at`

func (r *EntityRepository) Upsert(ctx context.Context, ns *namespace.Namespace, ent *entity.Entity) (string, error) {
	now := time.Now().UTC()
	ent.UpdatedAt = now

	// Check if entity already exists
	existing, err := r.GetByURN(ctx, ns, ent.URN)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("check existing entity: %w", err)
	}

	if existing.ID != "" {
		// Update: set valid_to on old row, insert new row
		_, err := r.client.ExecContext(ctx,
			`UPDATE entities SET updated_at = $1, properties = $2, name = $3, description = $4, source = $5, type = $6
			 WHERE id = $7`,
			now, JSONMap(ent.Properties), ent.Name, ent.Description, ent.Source, string(ent.Type), existing.ID)
		if err != nil {
			return "", fmt.Errorf("update entity: %w", err)
		}
		return existing.ID, nil
	}

	// Insert new entity
	var id string
	err = r.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx,
			`INSERT INTO entities (namespace_id, urn, type, name, description, properties, source, valid_from, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 RETURNING id`,
			ns.ID, ent.URN, string(ent.Type), ent.Name, ent.Description,
			JSONMap(ent.Properties), ent.Source, now, now, now,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("insert entity: %w", err)
	}
	ent.CreatedAt = now
	ent.ValidFrom = now
	return id, nil
}

func (r *EntityRepository) GetByURN(ctx context.Context, ns *namespace.Namespace, urn string) (entity.Entity, error) {
	q := fmt.Sprintf(`SELECT %s FROM entities WHERE namespace_id = $1 AND urn = $2 AND valid_to IS NULL LIMIT 1`, entityColumns)
	var m entityModel
	if err := r.client.GetContext(ctx, &m, q, ns.ID, urn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Entity{}, sql.ErrNoRows
		}
		return entity.Entity{}, fmt.Errorf("get entity by URN: %w", err)
	}
	return m.toEntity(), nil
}

func (r *EntityRepository) GetByID(ctx context.Context, id string) (entity.Entity, error) {
	q := fmt.Sprintf(`SELECT %s FROM entities WHERE id = $1 AND valid_to IS NULL LIMIT 1`, entityColumns)
	var m entityModel
	if err := r.client.GetContext(ctx, &m, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Entity{}, sql.ErrNoRows
		}
		return entity.Entity{}, fmt.Errorf("get entity by ID: %w", err)
	}
	return m.toEntity(), nil
}

func (r *EntityRepository) GetAll(ctx context.Context, ns *namespace.Namespace, flt entity.Filter) ([]entity.Entity, error) {
	builder := sq.Select(entityColumns).
		From("entities").
		Where(sq.Eq{"namespace_id": ns.ID}).
		Where("valid_to IS NULL").
		PlaceholderFormat(sq.Dollar)

	builder = applyEntityFilter(builder, flt)

	limit := flt.Size
	if limit <= 0 {
		limit = 50
	}
	builder = builder.Limit(uint64(limit)).Offset(uint64(flt.Offset))
	builder = builder.OrderBy("updated_at DESC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var models []entityModel
	if err := r.client.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("get entities: %w", err)
	}

	result := make([]entity.Entity, len(models))
	for i, m := range models {
		result[i] = m.toEntity()
	}
	return result, nil
}

func (r *EntityRepository) GetCount(ctx context.Context, ns *namespace.Namespace, flt entity.Filter) (int, error) {
	builder := sq.Select("count(1)").
		From("entities").
		Where(sq.Eq{"namespace_id": ns.ID}).
		Where("valid_to IS NULL").
		PlaceholderFormat(sq.Dollar)

	builder = applyEntityFilter(builder, flt)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var total int
	if err := r.client.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("count entities: %w", err)
	}
	return total, nil
}

func (r *EntityRepository) GetTypes(ctx context.Context, ns *namespace.Namespace) (map[entity.Type]int, error) {
	query := `SELECT type, count(1) as count FROM entities WHERE namespace_id = $1 AND valid_to IS NULL GROUP BY type`

	type typeCount struct {
		Type  string `db:"type"`
		Count int    `db:"count"`
	}
	var rows []typeCount
	if err := r.client.SelectContext(ctx, &rows, query, ns.ID); err != nil {
		return nil, fmt.Errorf("get entity types: %w", err)
	}

	result := make(map[entity.Type]int, len(rows))
	for _, r := range rows {
		result[entity.Type(r.Type)] = r.Count
	}
	return result, nil
}

func (r *EntityRepository) Delete(ctx context.Context, ns *namespace.Namespace, urn string) error {
	// Soft delete: set valid_to
	res, err := r.client.ExecContext(ctx,
		`UPDATE entities SET valid_to = now(), updated_at = now() WHERE namespace_id = $1 AND urn = $2 AND valid_to IS NULL`,
		ns.ID, urn)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func applyEntityFilter(builder sq.SelectBuilder, flt entity.Filter) sq.SelectBuilder {
	if len(flt.Types) > 0 {
		types := make([]string, len(flt.Types))
		for i, t := range flt.Types {
			types[i] = string(t)
		}
		builder = builder.Where(sq.Eq{"type": types})
	}
	if flt.Source != "" {
		builder = builder.Where(sq.Eq{"source": flt.Source})
	}
	if flt.Query != "" {
		builder = builder.Where(sq.Or{
			sq.ILike{"name": "%" + flt.Query + "%"},
			sq.ILike{"urn": "%" + flt.Query + "%"},
		})
	}
	return builder
}

// entityModel maps to the entities table.
type entityModel struct {
	ID          string     `db:"id"`
	NamespaceID string     `db:"namespace_id"`
	URN         string     `db:"urn"`
	Type        string     `db:"type"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Properties  JSONMap    `db:"properties"`
	Source      string     `db:"source"`
	ValidFrom   time.Time  `db:"valid_from"`
	ValidTo     *time.Time `db:"valid_to"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func (m entityModel) toEntity() entity.Entity {
	return entity.Entity{
		ID:          m.ID,
		NamespaceID: m.NamespaceID,
		URN:         m.URN,
		Type:        entity.Type(m.Type),
		Name:        m.Name,
		Description: m.Description,
		Properties:  m.Properties,
		Source:      m.Source,
		ValidFrom:   m.ValidFrom,
		ValidTo:     m.ValidTo,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}


