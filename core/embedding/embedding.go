package embedding

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Embedding is a text fragment with a vector embedding for semantic search.
// Derived from entities and documents — a rebuildable search index.
type Embedding struct {
	ID          string    `json:"id"`
	NamespaceID string    `json:"namespace_id"`
	EntityURN   string    `json:"entity_urn"`
	ContentID   string    `json:"content_id,omitempty"`
	ContentType string    `json:"content_type,omitempty"` // "entity" or "document"
	Content     string    `json:"content"`
	Context     string    `json:"context"`
	Vector      []float32 `json:"vector,omitempty"`
	Position    int       `json:"position"`
	Heading     string    `json:"heading,omitempty"`
	TokenCount  int       `json:"token_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Repository defines storage operations for the embedding index.
type Repository interface {
	UpsertBatch(ctx context.Context, ns *namespace.Namespace, embeddings []Embedding) error
	DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error
	DeleteByContentID(ctx context.Context, ns *namespace.Namespace, contentID string) error
	Search(ctx context.Context, ns *namespace.Namespace, vector []float32, limit int) ([]Embedding, error)
}
