package entity

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Type is an open type system — any non-empty string is valid.
type Type string

func (t Type) String() string { return string(t) }

func (t Type) IsValid() bool { return t != "" }

// Well-known types for convenience. Not an exhaustive list.
const (
	TypeTable        Type = "table"
	TypeJob          Type = "job"
	TypeDashboard    Type = "dashboard"
	TypeTopic        Type = "topic"
	TypeFeatureTable Type = "feature_table"
	TypeApplication  Type = "application"
	TypeModel        Type = "model"
	TypeQuery        Type = "query"
	TypeMetric       Type = "metric"
	TypeExperiment   Type = "experiment"
)

// Entity is the core domain object — anything worth naming in the
// organization's knowledge graph.
type Entity struct {
	ID          string                 `json:"id"`
	NamespaceID string                 `json:"namespace_id"`
	URN         string                 `json:"urn"`
	Type        Type                   `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Source      string                 `json:"source,omitempty"`
	ValidFrom   time.Time              `json:"valid_from"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// IsCurrent returns true if this entity represents current state.
func (e Entity) IsCurrent() bool { return e.ValidTo == nil }

// Repository defines storage operations for entities.
type Repository interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, ent *Entity) (string, error)
	GetByURN(ctx context.Context, ns *namespace.Namespace, urn string) (Entity, error)
	GetByID(ctx context.Context, id string) (Entity, error)
	GetAll(ctx context.Context, ns *namespace.Namespace, filter Filter) ([]Entity, error)
	GetCount(ctx context.Context, ns *namespace.Namespace, filter Filter) (int, error)
	GetTypes(ctx context.Context, ns *namespace.Namespace) (map[Type]int, error)
	Delete(ctx context.Context, ns *namespace.Namespace, urn string) error
}

// Filter for querying entities.
type Filter struct {
	Types  []Type
	Source string
	Size   int
	Offset int
	Query  string
}
