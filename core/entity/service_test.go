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

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:x", Type: TypeJob, Name: "x"})

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

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeJob, Name: "b"})

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

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeJob, Name: "c"})

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

// mockEdgeRepo is a simple in-memory edge repository for testing.
type mockEdgeRepo struct {
	edges               []Edge
	downstreamEdges     []Edge
	lastDownstreamDepth int
}

func (m *mockEdgeRepo) Upsert(_ context.Context, _ *namespace.Namespace, e *Edge) error {
	m.edges = append(m.edges, *e)
	return nil
}

func (m *mockEdgeRepo) GetBySource(_ context.Context, _ *namespace.Namespace, urn string, _ EdgeFilter) ([]Edge, error) {
	var result []Edge
	for _, e := range m.edges {
		if e.SourceURN == urn {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEdgeRepo) GetByTarget(_ context.Context, _ *namespace.Namespace, urn string, _ EdgeFilter) ([]Edge, error) {
	var result []Edge
	for _, e := range m.edges {
		if e.TargetURN == urn {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEdgeRepo) GetDownstream(_ context.Context, _ *namespace.Namespace, _ string, depth int) ([]Edge, error) {
	m.lastDownstreamDepth = depth
	return m.downstreamEdges, nil
}

func (m *mockEdgeRepo) GetUpstream(_ context.Context, _ *namespace.Namespace, _ string, _ int) ([]Edge, error) {
	return nil, nil
}

func (m *mockEdgeRepo) GetBidirectional(_ context.Context, _ *namespace.Namespace, urn string, depth int) ([]Edge, error) {
	// BFS traversal up to depth hops in both directions.
	type frontier struct {
		urn   string
		level int
	}
	visited := map[string]bool{urn: true}
	queue := []frontier{{urn: urn, level: 0}}
	var result []Edge
	seen := map[string]bool{} // dedup edges by source+target+type

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.level >= depth {
			continue
		}
		for _, e := range m.edges {
			key := e.SourceURN + "|" + e.TargetURN + "|" + e.Type
			var neighbor string
			if e.SourceURN == cur.urn {
				neighbor = e.TargetURN
			} else if e.TargetURN == cur.urn {
				neighbor = e.SourceURN
			} else {
				continue
			}
			if !seen[key] {
				seen[key] = true
				result = append(result, e)
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, frontier{urn: neighbor, level: cur.level + 1})
			}
		}
	}
	return result, nil
}

func (m *mockEdgeRepo) Delete(_ context.Context, _ *namespace.Namespace, _, _, _ string) error {
	return nil
}

func (m *mockEdgeRepo) DeleteByURN(_ context.Context, _ *namespace.Namespace, _ string) error {
	return nil
}

func TestService_GetContext_DefaultDepth(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	// A -> B -> C (linear chain)
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeTable, Name: "c"})
	edges.edges = []Edge{
		{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
		{SourceURN: "urn:b", TargetURN: "urn:c", Type: "lineage"},
	}

	// depth=0 should default to 1 (only direct neighbors of B)
	cg, err := svc.GetContext(ctx, ns, "urn:b", 0)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}
	if len(cg.Edges) != 2 {
		t.Errorf("expected 2 edges at depth 1, got %d", len(cg.Edges))
	}
	if len(cg.Related) != 2 {
		t.Errorf("expected 2 related entities, got %d", len(cg.Related))
	}
}

func TestService_GetContext_MultiHop(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	// A -> B -> C -> D
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeTable, Name: "c"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:d", Type: TypeTable, Name: "d"})
	edges.edges = []Edge{
		{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
		{SourceURN: "urn:b", TargetURN: "urn:c", Type: "lineage"},
		{SourceURN: "urn:c", TargetURN: "urn:d", Type: "lineage"},
	}

	// depth=2 from B should reach A, C, and D
	cg, err := svc.GetContext(ctx, ns, "urn:b", 2)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}
	if len(cg.Edges) != 3 {
		t.Errorf("expected 3 edges at depth 2, got %d", len(cg.Edges))
	}
	if len(cg.Related) != 3 {
		t.Errorf("expected 3 related entities (a, c, d), got %d", len(cg.Related))
	}
}

func TestService_GetContext_MaxDepthCap(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})

	// depth=10 should be capped to maxContextDepth (5), not error
	cg, err := svc.GetContext(ctx, ns, "urn:a", 10)
	if err != nil {
		t.Fatalf("GetContext with large depth should not error: %v", err)
	}
	if cg.Entity.URN != "urn:a" {
		t.Errorf("expected entity urn:a, got %s", cg.Entity.URN)
	}
}

func TestService_GetContext_NilEdges(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil) // no edge repo
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})

	cg, err := svc.GetContext(ctx, ns, "urn:a", 2)
	if err != nil {
		t.Fatalf("GetContext with nil edges should not error: %v", err)
	}
	if len(cg.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(cg.Edges))
	}
	if len(cg.Related) != 0 {
		t.Errorf("expected 0 related, got %d", len(cg.Related))
	}
}

func TestService_GetContext_CycleHandling(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	// Cycle: A -> B -> C -> A
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:a", Type: TypeTable, Name: "a"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:b", Type: TypeTable, Name: "b"})
	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:c", Type: TypeTable, Name: "c"})
	edges.edges = []Edge{
		{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
		{SourceURN: "urn:b", TargetURN: "urn:c", Type: "lineage"},
		{SourceURN: "urn:c", TargetURN: "urn:a", Type: "lineage"},
	}

	// depth=3 should not infinite loop
	cg, err := svc.GetContext(ctx, ns, "urn:a", 3)
	if err != nil {
		t.Fatalf("GetContext with cycle should not error: %v", err)
	}
	if len(cg.Edges) != 3 {
		t.Errorf("expected 3 edges in cycle, got %d", len(cg.Edges))
	}
	if len(cg.Related) != 2 {
		t.Errorf("expected 2 related entities (b, c), got %d", len(cg.Related))
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

func TestService_GetByID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	_, _ = svc.Upsert(ctx, ns, &Entity{URN: "urn:table:orders", Type: TypeTable, Name: "orders", Source: "bigquery"})

	got, err := svc.GetByID(ctx, "urn:table:orders")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "orders" {
		t.Errorf("expected name 'orders', got %q", got.Name)
	}
	if got.Source != "bigquery" {
		t.Errorf("expected source 'bigquery', got %q", got.Source)
	}

	// non-existent ID should error
	_, err = svc.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
}

func TestService_UpsertWithEdges(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	ent := &Entity{URN: "urn:table:main", Type: TypeTable, Name: "main"}
	upstreams := []string{"urn:table:source1", "urn:table:source2"}
	downstreams := []string{"urn:table:sink1"}

	id, err := svc.UpsertWithEdges(ctx, ns, ent, upstreams, downstreams)
	if err != nil {
		t.Fatalf("UpsertWithEdges failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	// Should have 2 upstream edges + 1 downstream edge = 3 total
	if len(edges.edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(edges.edges))
	}

	// Check upstream edges: source -> main
	if edges.edges[0].SourceURN != "urn:table:source1" || edges.edges[0].TargetURN != "urn:table:main" {
		t.Errorf("upstream edge 0: got source=%q target=%q", edges.edges[0].SourceURN, edges.edges[0].TargetURN)
	}
	if edges.edges[1].SourceURN != "urn:table:source2" || edges.edges[1].TargetURN != "urn:table:main" {
		t.Errorf("upstream edge 1: got source=%q target=%q", edges.edges[1].SourceURN, edges.edges[1].TargetURN)
	}

	// Check downstream edge: main -> sink
	if edges.edges[2].SourceURN != "urn:table:main" || edges.edges[2].TargetURN != "urn:table:sink1" {
		t.Errorf("downstream edge: got source=%q target=%q", edges.edges[2].SourceURN, edges.edges[2].TargetURN)
	}

	// All edges should be lineage type
	for i, e := range edges.edges {
		if e.Type != "lineage" {
			t.Errorf("edge %d: expected type 'lineage', got %q", i, e.Type)
		}
	}
}

func TestService_GetImpact(t *testing.T) {
	repo := newMockRepo()
	edges := &mockEdgeRepo{
		downstreamEdges: []Edge{
			{SourceURN: "urn:a", TargetURN: "urn:b", Type: "lineage"},
			{SourceURN: "urn:b", TargetURN: "urn:c", Type: "lineage"},
		},
	}
	svc := NewService(repo, edges, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	// default depth (0 -> 3)
	result, err := svc.GetImpact(ctx, ns, "urn:a", 0)
	if err != nil {
		t.Fatalf("GetImpact failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 downstream edges, got %d", len(result))
	}
	if edges.lastDownstreamDepth != 3 {
		t.Errorf("expected default depth 3, got %d", edges.lastDownstreamDepth)
	}

	// custom depth
	_, err = svc.GetImpact(ctx, ns, "urn:a", 5)
	if err != nil {
		t.Fatalf("GetImpact with custom depth failed: %v", err)
	}
	if edges.lastDownstreamDepth != 5 {
		t.Errorf("expected depth 5, got %d", edges.lastDownstreamDepth)
	}
}

func TestService_GetImpact_NilEdges(t *testing.T) {
	svc := NewService(newMockRepo(), nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	result, err := svc.GetImpact(ctx, ns, "urn:a", 2)
	if err != nil {
		t.Fatalf("GetImpact with nil edges should not error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestService_Suggest_NilSearchRepo(t *testing.T) {
	svc := NewService(newMockRepo(), nil, nil)
	ctx := context.Background()
	ns := namespace.DefaultNamespace

	result, err := svc.Suggest(ctx, ns, "test", 10)
	if err != nil {
		t.Fatalf("Suggest with nil search repo should not error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// mockSearchRepo is a simple in-memory search repository for testing.
type mockSearchRepo struct {
	results     []SearchResult
	suggestions []string
}

func (m *mockSearchRepo) Search(_ context.Context, _ SearchConfig) ([]SearchResult, error) {
	return m.results, nil
}

func (m *mockSearchRepo) Suggest(_ context.Context, _ *namespace.Namespace, _ string, _ int) ([]string, error) {
	return m.suggestions, nil
}

func TestService_Search_WithKeyword(t *testing.T) {
	search := &mockSearchRepo{
		results: []SearchResult{
			{URN: "urn:table:orders", Name: "orders", Type: "table"},
			{URN: "urn:table:users", Name: "users", Type: "table"},
		},
	}
	svc := NewService(newMockRepo(), nil, search)
	ctx := context.Background()

	results, err := svc.Search(ctx, SearchConfig{Text: "orders", Mode: SearchModeKeyword})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].URN != "urn:table:orders" {
		t.Errorf("expected first result URN 'urn:table:orders', got %q", results[0].URN)
	}
}
