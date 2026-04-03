package document

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/raystack/compass/core/namespace"
)

type mockRepo struct {
	documents map[string]Document
}

func newMockRepo() *mockRepo {
	return &mockRepo{documents: make(map[string]Document)}
}

func (m *mockRepo) Upsert(_ context.Context, _ *namespace.Namespace, doc *Document) (string, error) {
	id := doc.EntityURN + "/" + doc.Title
	doc.ID = id
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	m.documents[id] = *doc
	return id, nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (Document, error) {
	if d, ok := m.documents[id]; ok {
		return d, nil
	}
	return Document{}, sql.ErrNoRows
}

func (m *mockRepo) GetByEntityURN(_ context.Context, _ *namespace.Namespace, entityURN string) ([]Document, error) {
	var result []Document
	for _, d := range m.documents {
		if d.EntityURN == entityURN {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockRepo) GetAll(_ context.Context, _ *namespace.Namespace, _ Filter) ([]Document, error) {
	var result []Document
	for _, d := range m.documents {
		result = append(result, d)
	}
	return result, nil
}

func (m *mockRepo) Delete(_ context.Context, _ *namespace.Namespace, id string) error {
	if _, ok := m.documents[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.documents, id)
	return nil
}

func (m *mockRepo) DeleteByEntityURN(_ context.Context, _ *namespace.Namespace, entityURN string) error {
	for id, d := range m.documents {
		if d.EntityURN == entityURN {
			delete(m.documents, id)
		}
	}
	return nil
}

func TestService_UpsertAndGet(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	doc := &Document{
		EntityURN: "urn:table:orders",
		Title:     "Orders Table Runbook",
		Body:      "## Overview\nThe orders table tracks customer purchases.",
		Source:    "confluence",
		SourceID:  "page-123",
	}

	id, err := svc.Upsert(ctx, ns, doc)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	got, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Title != "Orders Table Runbook" {
		t.Errorf("expected title 'Orders Table Runbook', got %q", got.Title)
	}
	if got.Format != "markdown" {
		t.Errorf("expected default format 'markdown', got %q", got.Format)
	}
}

func TestService_GetByEntityURN(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:orders", Title: "Runbook", Body: "content"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:orders", Title: "Schema Docs", Body: "content"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:users", Title: "Users Docs", Body: "content"})

	docs, err := svc.GetByEntityURN(ctx, ns, "urn:table:orders")
	if err != nil {
		t.Fatalf("GetByEntityURN failed: %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	id, _ := svc.Upsert(ctx, ns, &Document{EntityURN: "urn:x", Title: "Doc", Body: "body"})

	if err := svc.Delete(ctx, ns, id); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := svc.GetByID(ctx, id)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
