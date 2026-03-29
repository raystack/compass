package entity

import (
	"context"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// Chunk is a text fragment with a vector embedding for semantic search.
// Derived from entities — an indexing mechanism, not knowledge.
type Chunk struct {
	ID          string    `json:"id"`
	NamespaceID string    `json:"namespace_id"`
	EntityURN   string    `json:"entity_urn"`
	Content     string    `json:"content"`
	Context     string    `json:"context"`
	Embedding   []float32 `json:"embedding,omitempty"`
	Position    int       `json:"position"`
	Heading     string    `json:"heading,omitempty"`
	TokenCount  int       `json:"token_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChunkRepository defines storage operations for the chunk index.
type ChunkRepository interface {
	UpsertBatch(ctx context.Context, ns *namespace.Namespace, chunks []Chunk) error
	DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error
	Search(ctx context.Context, ns *namespace.Namespace, embedding []float32, limit int) ([]Chunk, error)
}
