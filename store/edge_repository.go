package store

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

type EdgeRepository struct {
	client *Client
}

func NewEdgeRepository(client *Client) (*EdgeRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &EdgeRepository{client: client}, nil
}

var edgeColumns = `id, namespace_id, source_urn, target_urn, type, properties, valid_from, valid_to, source, created_at`

func (r *EdgeRepository) Upsert(ctx context.Context, ns *namespace.Namespace, e *entity.Edge) error {
	query, args, err := sq.Insert("edges").
		Columns("namespace_id", "source_urn", "target_urn", "type", "properties", "source").
		Values(ns.ID, e.SourceURN, e.TargetURN, e.Type, JSONMap(e.Properties), e.Source).
		Suffix(`ON CONFLICT (namespace_id, source_urn, target_urn, type, valid_from)
			DO UPDATE SET properties = EXCLUDED.properties, source = EXCLUDED.source`).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build upsert edge: %w", err)
	}
	_, err = r.client.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("upsert edge: %w", err)
	}
	return nil
}

func (r *EdgeRepository) GetBySource(ctx context.Context, ns *namespace.Namespace, urn string, filter entity.EdgeFilter) ([]entity.Edge, error) {
	builder := sq.Select(edgeColumns).From("edges").
		Where(sq.Eq{"namespace_id": ns.ID, "source_urn": urn}).
		PlaceholderFormat(sq.Dollar)
	builder = applyEdgeFilter(builder, filter)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	return r.queryEdges(ctx, query, args...)
}

func (r *EdgeRepository) GetByTarget(ctx context.Context, ns *namespace.Namespace, urn string, filter entity.EdgeFilter) ([]entity.Edge, error) {
	builder := sq.Select(edgeColumns).From("edges").
		Where(sq.Eq{"namespace_id": ns.ID, "target_urn": urn}).
		PlaceholderFormat(sq.Dollar)
	builder = applyEdgeFilter(builder, filter)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	return r.queryEdges(ctx, query, args...)
}

func (r *EdgeRepository) GetDownstream(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error) {
	return r.traverse(ctx, ns, urn, depth, "downstream")
}

func (r *EdgeRepository) GetUpstream(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error) {
	return r.traverse(ctx, ns, urn, depth, "upstream")
}

func (r *EdgeRepository) Delete(ctx context.Context, ns *namespace.Namespace, sourceURN, targetURN, edgeType string) error {
	_, err := r.client.ExecContext(ctx,
		`UPDATE edges SET valid_to = now() WHERE namespace_id = $1 AND source_urn = $2 AND target_urn = $3 AND type = $4 AND valid_to IS NULL`,
		ns.ID, sourceURN, targetURN, edgeType)
	return err
}

func (r *EdgeRepository) DeleteByURN(ctx context.Context, ns *namespace.Namespace, urn string) error {
	_, err := r.client.ExecContext(ctx,
		`UPDATE edges SET valid_to = now() WHERE namespace_id = $1 AND (source_urn = $2 OR target_urn = $2) AND valid_to IS NULL`,
		ns.ID, urn)
	return err
}

func (r *EdgeRepository) traverse(ctx context.Context, ns *namespace.Namespace, urn string, depth int, direction string) ([]entity.Edge, error) {
	if depth <= 0 {
		depth = 3
	}

	var seedCol, joinCol string
	if direction == "downstream" {
		seedCol, joinCol = "source_urn", "target_urn"
	} else {
		seedCol, joinCol = "target_urn", "source_urn"
	}

	query := fmt.Sprintf(`
		WITH RECURSIVE graph(source_urn, target_urn, type, properties, depth, path) AS (
			SELECT source_urn, target_urn, type, properties, 1, ARRAY[%s]
			FROM edges
			WHERE namespace_id = $1 AND %s = $2 AND valid_to IS NULL
		UNION ALL
			SELECT e.source_urn, e.target_urn, e.type, e.properties, g.depth + 1, g.path || e.%s
			FROM edges e
			JOIN graph g ON e.%s = g.%s
			WHERE e.%s <> ALL(g.path) AND e.valid_to IS NULL AND g.depth < $3
		)
		SELECT source_urn, target_urn, type, properties FROM graph`,
		seedCol, seedCol, seedCol, joinCol, joinCol, seedCol)

	var models []edgeModel
	if err := r.client.SelectContext(ctx, &models, query, ns.ID, urn, depth); err != nil {
		return nil, fmt.Errorf("traverse %s: %w", direction, err)
	}
	return toEdgeList(models), nil
}

func (r *EdgeRepository) queryEdges(ctx context.Context, query string, args ...interface{}) ([]entity.Edge, error) {
	var models []edgeModel
	if err := r.client.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("query edges: %w", err)
	}
	return toEdgeList(models), nil
}

func applyEdgeFilter(builder sq.SelectBuilder, filter entity.EdgeFilter) sq.SelectBuilder {
	if filter.Current {
		builder = builder.Where("valid_to IS NULL")
	}
	if len(filter.Types) > 0 {
		builder = builder.Where(sq.Eq{"type": filter.Types})
	}
	return builder
}

type edgeModel struct {
	ID          string     `db:"id"`
	NamespaceID string     `db:"namespace_id"`
	SourceURN   string     `db:"source_urn"`
	TargetURN   string     `db:"target_urn"`
	Type        string     `db:"type"`
	Properties  JSONMap    `db:"properties"`
	ValidFrom   time.Time  `db:"valid_from"`
	ValidTo     *time.Time `db:"valid_to"`
	Source      string     `db:"source"`
	CreatedAt   time.Time  `db:"created_at"`
}

func toEdgeList(models []edgeModel) []entity.Edge {
	edges := make([]entity.Edge, len(models))
	for i, m := range models {
		edges[i] = entity.Edge{
			ID:          m.ID,
			NamespaceID: m.NamespaceID,
			SourceURN:   m.SourceURN,
			TargetURN:   m.TargetURN,
			Type:        m.Type,
			Properties:  m.Properties,
			ValidFrom:   m.ValidFrom,
			ValidTo:     m.ValidTo,
			Source:      m.Source,
			CreatedAt:   m.CreatedAt,
		}
	}
	return edges
}

var _ = strings.Join // keep import
