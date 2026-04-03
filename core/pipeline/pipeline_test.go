package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/embedding"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

type mockProvider struct {
	dims int
}

func (m *mockProvider) Name() string      { return "mock" }
func (m *mockProvider) Dimensions() int   { return m.dims }
func (m *mockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return make([]float32, m.dims), nil
}
func (m *mockProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, m.dims)
	}
	return result, nil
}

type mockEmbeddingRepo struct {
	mu         sync.Mutex
	embeddings []embedding.Embedding
}

func (m *mockEmbeddingRepo) UpsertBatch(_ context.Context, _ *namespace.Namespace, embs []embedding.Embedding) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embeddings = append(m.embeddings, embs...)
	return nil
}

func (m *mockEmbeddingRepo) DeleteByEntityURN(_ context.Context, _ *namespace.Namespace, _ string) error {
	return nil
}

func (m *mockEmbeddingRepo) DeleteByContentID(_ context.Context, _ *namespace.Namespace, _ string) error {
	return nil
}

func (m *mockEmbeddingRepo) Search(_ context.Context, _ *namespace.Namespace, _ []float32, _ int) ([]embedding.Embedding, error) {
	return nil, nil
}

func (m *mockEmbeddingRepo) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.embeddings)
}

func TestPipeline_EnqueueEntity(t *testing.T) {
	repo := &mockEmbeddingRepo{}
	provider := &mockProvider{dims: 768}
	p := New(repo, provider, WithWorkers(1), WithQueueSize(10))

	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)

	ns := namespace.DefaultNamespace
	ent := &entity.Entity{
		ID:          "ent-1",
		URN:         "urn:table:orders",
		Type:        entity.TypeTable,
		Name:        "orders",
		Description: "Customer orders",
	}

	if err := p.EnqueueEntity(ctx, ns, ent); err != nil {
		t.Fatalf("EnqueueEntity failed: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	cancel()
	p.Stop()

	if repo.count() == 0 {
		t.Fatal("expected embeddings to be created")
	}

	emb := repo.embeddings[0]
	if emb.EntityURN != "urn:table:orders" {
		t.Errorf("unexpected entity URN: %s", emb.EntityURN)
	}
	if emb.ContentType != "entity" {
		t.Errorf("unexpected content type: %s", emb.ContentType)
	}
	if len(emb.Vector) != 768 {
		t.Errorf("expected 768-dim vector, got %d", len(emb.Vector))
	}
}

func TestPipeline_EnqueueDocument(t *testing.T) {
	repo := &mockEmbeddingRepo{}
	provider := &mockProvider{dims: 768}
	p := New(repo, provider, WithWorkers(1), WithQueueSize(10))

	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)

	ns := namespace.DefaultNamespace
	doc := &document.Document{
		ID:        "doc-1",
		EntityURN: "urn:table:orders",
		Title:     "Orders Runbook",
		Body: `## Overview
The orders table is refreshed daily.

## Recovery
If the pipeline fails, restart the DAG.`,
	}

	if err := p.EnqueueDocument(ctx, ns, doc); err != nil {
		t.Fatalf("EnqueueDocument failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	cancel()
	p.Stop()

	if repo.count() < 2 {
		t.Fatalf("expected at least 2 chunks from document, got %d", repo.count())
	}

	for _, emb := range repo.embeddings {
		if emb.ContentType != "document" {
			t.Errorf("unexpected content type: %s", emb.ContentType)
		}
		if emb.EntityURN != "urn:table:orders" {
			t.Errorf("unexpected entity URN: %s", emb.EntityURN)
		}
	}
}

func TestPipeline_QueueFull(t *testing.T) {
	repo := &mockEmbeddingRepo{}
	provider := &mockProvider{dims: 3}
	p := New(repo, provider, WithWorkers(0), WithQueueSize(1)) // 0 workers = no processing

	// Don't start — no workers consuming
	ns := namespace.DefaultNamespace

	// First enqueue should succeed (fills the buffer)
	_ = p.EnqueueEntity(context.Background(), ns, &entity.Entity{URN: "urn:1", Name: "a", Type: "t"})

	// Second should be dropped (queue full), but no error returned
	err := p.EnqueueEntity(context.Background(), ns, &entity.Entity{URN: "urn:2", Name: "b", Type: "t"})
	if err != nil {
		t.Fatalf("expected no error on queue full, got: %v", err)
	}
}
