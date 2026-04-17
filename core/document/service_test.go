package document

import (
	"context"
	"database/sql"
	"fmt"
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
	if doc.EntityURN == "" {
		return "", fmt.Errorf("entity_urn is required")
	}
	if doc.Title == "" {
		return "", fmt.Errorf("title is required")
	}
	if doc.Body == "" {
		return "", fmt.Errorf("body is required")
	}
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

func TestService_GetAll(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:a", Title: "Doc A", Body: "body a"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:b", Title: "Doc B", Body: "body b"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:c", Title: "Doc C", Body: "body c"})

	docs, err := svc.GetAll(ctx, ns, Filter{})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("expected 3 documents, got %d", len(docs))
	}
}

func TestService_DeleteByEntityURN(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:orders", Title: "Runbook", Body: "content"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:orders", Title: "Schema", Body: "content"})
	_, _ = svc.Upsert(ctx, ns, &Document{EntityURN: "urn:table:users", Title: "Users Doc", Body: "content"})

	err := svc.DeleteByEntityURN(ctx, ns, "urn:table:orders")
	if err != nil {
		t.Fatalf("DeleteByEntityURN failed: %v", err)
	}

	// orders docs should be gone
	docs, err := svc.GetByEntityURN(ctx, ns, "urn:table:orders")
	if err != nil {
		t.Fatalf("GetByEntityURN failed: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("expected 0 documents for orders, got %d", len(docs))
	}

	// users doc should remain
	docs, err = svc.GetByEntityURN(ctx, ns, "urn:table:users")
	if err != nil {
		t.Fatalf("GetByEntityURN failed: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 document for users, got %d", len(docs))
	}
}

func TestService_Upsert_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	tests := []struct {
		name string
		doc  *Document
	}{
		{
			name: "missing entity_urn",
			doc:  &Document{Title: "Some Title", Body: "some body"},
		},
		{
			name: "missing title",
			doc:  &Document{EntityURN: "urn:table:x", Body: "some body"},
		},
		{
			name: "missing body",
			doc:  &Document{EntityURN: "urn:table:x", Title: "Some Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Upsert(ctx, ns, tt.doc)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}
