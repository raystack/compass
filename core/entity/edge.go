package entity

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Edge is a typed, directed, temporal relationship between two entities.
type Edge struct {
	ID          string                 `json:"id"`
	NamespaceID string                 `json:"namespace_id"`
	SourceURN   string                 `json:"source_urn"`
	TargetURN   string                 `json:"target_urn"`
	Type        string                 `json:"type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	ValidFrom   time.Time              `json:"valid_from"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	Source      string                 `json:"source,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// EdgeFilter for querying edges.
type EdgeFilter struct {
	Types   []string
	Current bool // only current edges (valid_to IS NULL)
}

// EdgeRepository defines storage operations for edges.
type EdgeRepository interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, e *Edge) error
	GetBySource(ctx context.Context, ns *namespace.Namespace, urn string, filter EdgeFilter) ([]Edge, error)
	GetByTarget(ctx context.Context, ns *namespace.Namespace, urn string, filter EdgeFilter) ([]Edge, error)
	GetDownstream(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]Edge, error)
	GetUpstream(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]Edge, error)
	GetBidirectional(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]Edge, error)
	Delete(ctx context.Context, ns *namespace.Namespace, sourceURN, targetURN, edgeType string) error
	DeleteByURN(ctx context.Context, ns *namespace.Namespace, urn string) error
}
