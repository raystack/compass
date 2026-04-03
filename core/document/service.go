package document

import (
	"context"
	"fmt"

	"github.com/raystack/compass/core/namespace"
)

// EmbeddingPipeline enqueues documents for async embedding.
type EmbeddingPipeline interface {
	EnqueueDocument(ctx context.Context, ns *namespace.Namespace, doc *Document) error
}

// Service orchestrates document operations.
type Service struct {
	repo     Repository
	pipeline EmbeddingPipeline
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// WithPipeline enables async embedding on document upsert.
func (s *Service) WithPipeline(p EmbeddingPipeline) {
	s.pipeline = p
}

func (s *Service) Upsert(ctx context.Context, ns *namespace.Namespace, doc *Document) (string, error) {
	if doc.Format == "" {
		doc.Format = "markdown"
	}
	id, err := s.repo.Upsert(ctx, ns, doc)
	if err != nil {
		return "", fmt.Errorf("upsert document: %w", err)
	}
	doc.ID = id

	if s.pipeline != nil {
		_ = s.pipeline.EnqueueDocument(ctx, ns, doc)
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Document, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]Document, error) {
	return s.repo.GetByEntityURN(ctx, ns, entityURN)
}

func (s *Service) GetAll(ctx context.Context, ns *namespace.Namespace, filter Filter) ([]Document, error) {
	return s.repo.GetAll(ctx, ns, filter)
}

func (s *Service) Delete(ctx context.Context, ns *namespace.Namespace, id string) error {
	return s.repo.Delete(ctx, ns, id)
}

func (s *Service) DeleteByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) error {
	return s.repo.DeleteByEntityURN(ctx, ns, entityURN)
}
