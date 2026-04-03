package document

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Document is a piece of knowledge linked to an entity.
// Compass is the source of truth for documents — they persist
// even if the original source system is unavailable.
type Document struct {
	ID          string                 `json:"id"`
	NamespaceID string                 `json:"namespace_id"`
	EntityURN   string                 `json:"entity_urn"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Format      string                 `json:"format,omitempty"`    // "markdown", "plaintext"
	Source      string                 `json:"source,omitempty"`    // "confluence", "github", "manual"
	SourceID    string                 `json:"source_id,omitempty"` // original doc ID for dedup
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Filter for querying documents.
type Filter struct {
	EntityURN string
	Source    string
	Size     int
	Offset   int
}

// Repository defines storage operations for documents.
type Repository interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, doc *Document) (string, error)
	GetByID(ctx context.Context, id string) (Document, error)
	GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]Document, error)
	GetAll(ctx context.Context, ns *namespace.Namespace, filter Filter) ([]Document, error)
	Delete(ctx context.Context, ns *namespace.Namespace, id string) error
	DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error
}
