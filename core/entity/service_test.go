package entity

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/raystack/compass/core/namespace"
)

// mockRepo is a simple in-memory entity repository for testing.
type mockRepo struct {
	entities map[string]Entity
}

func newMockRepo() *mockRepo {
	return &mockRepo{entities: make(map[string]Entity)}
}

func (m *mockRepo) Upsert(_ context.Context, _ *namespace.Namespace, ent *Entity) (string, error) {
	id := ent.URN
	ent.ID = id
	ent.CreatedAt = time.Now()
	ent.UpdatedAt = time.Now()
	m.entities[id] = *ent
	return id, nil
}

func (m *mockRepo) GetByURN(_ context.Context, _ *namespace.Namespace, urn string) (Entity, error) {
	if e, ok := m.entities[urn]; ok {
		return e, nil
	}
	return Entity{}, sql.ErrNoRows
}

func (m *mockRepo) GetByID(_ context.Context, id string) (Entity, error) {
	for _, e := range m.entities {
		if e.ID == id {
			return e, nil
		}
	}
	return Entity{}, sql.ErrNoRows
}

func (m *mockRepo) GetAll(_ context.Context, _ *namespace.Namespace, _ Filter) ([]Entity, error) {
	var result []Entity
	for _, e := range m.entities {
		result = append(result, e)
	}
	return result, nil
}

func (m *mockRepo) GetCount(_ context.Context, _ *namespace.Namespace, _ Filter) (int, error) {
	return len(m.entities), nil
}

func (m *mockRepo) GetTypes(_ context.Context, _ *namespace.Namespace) (map[Type]int, error) {
	types := make(map[Type]int)
	for _, e := range m.entities {
		types[e.Type]++
	}
	return types, nil
}

func (m *mockRepo) Delete(_ context.Context, _ *namespace.Namespace, urn string) error {
	if _, ok := m.entities[urn]; !ok {
		return sql.ErrNoRows
	}
	delete(m.entities, urn)
	return nil
}

func TestService_UpsertAndGet(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	ent := &Entity{
		URN:    "urn:table:test",
		Type:   TypeTable,
		Name:   "test_table",
		Source: "bigquery",
	}

	id, err := svc.Upsert(ctx, ns, ent)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	got, err := svc.GetByURN(ctx, ns, "urn:table:test")
	if err != nil {
		t.Fatalf("GetByURN failed: %v", err)
	}
	if got.Name != "test_table" {
		t.Errorf("expected name 'test_table', got %q", got.Name)
	}
	if got.Source != "bigquery" {
		t.Errorf("expected source 'bigquery', got %q", got.Source)
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	svc.Upsert(ctx, ns, &Entity{URN: "urn:x", Type: TypeJob, Name: "x"})

	err := svc.Delete(ctx, ns, "urn:x")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = svc.GetByURN(ctx, ns, "urn:x")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestService_GetAll(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeJob, Name: "b"})

	entities, count, err := svc.GetAll(ctx, ns, Filter{})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(entities))
	}
}

func TestService_GetTypes(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b"})
	svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeJob, Name: "c"})

	types, err := svc.GetTypes(ctx, ns)
	if err != nil {
		t.Fatalf("GetTypes failed: %v", err)
	}
	if types[TypeTable] != 2 {
		t.Errorf("expected 2 tables, got %d", types[TypeTable])
	}
	if types[TypeJob] != 1 {
		t.Errorf("expected 1 job, got %d", types[TypeJob])
	}
}

func TestService_Search_NilRepos(t *testing.T) {
	svc := NewService(newMockRepo(), nil, nil)
	results, err := svc.Search(context.Background(), SearchConfig{Text: "test"})
	if err != nil {
		t.Fatalf("Search with nil repos should not error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}
