package entity

import (
	"context"
	"fmt"

	"github.com/raystack/compass/core/namespace"
)

// HybridSearcher is an interface for hybrid search (keyword + semantic).
// Implemented by embedding.HybridSearch.
type HybridSearcher interface {
	Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error)
}

// EmbeddingPipeline enqueues entities for async embedding.
type EmbeddingPipeline interface {
	EnqueueEntity(ctx context.Context, ns *namespace.Namespace, ent *Entity) error
}

// Service orchestrates entity operations across repositories.
type Service struct {
	repo     Repository
	edges    EdgeRepository
	search   SearchRepository
	hybrid   HybridSearcher
	pipeline EmbeddingPipeline
}

func NewService(repo Repository, edges EdgeRepository, search SearchRepository) *Service {
	return &Service{repo: repo, edges: edges, search: search}
}

// WithHybridSearch enables semantic/hybrid search modes.
func (s *Service) WithHybridSearch(hs HybridSearcher) {
	s.hybrid = hs
}

// WithPipeline enables async embedding on entity upsert.
func (s *Service) WithPipeline(p EmbeddingPipeline) {
	s.pipeline = p
}

func (s *Service) Upsert(ctx context.Context, ns *namespace.Namespace, ent *Entity) (string, error) {
	id, err := s.repo.Upsert(ctx, ns, ent)
	if err != nil {
		return "", fmt.Errorf("upsert entity: %w", err)
	}
	ent.ID = id

	if s.pipeline != nil {
		_ = s.pipeline.EnqueueEntity(ctx, ns, ent)
	}

	return id, nil
}

func (s *Service) UpsertWithEdges(ctx context.Context, ns *namespace.Namespace, ent *Entity, upstreams, downstreams []string) (string, error) {
	id, err := s.Upsert(ctx, ns, ent)
	if err != nil {
		return "", err
	}

	if s.edges != nil {
		for _, us := range upstreams {
			_ = s.edges.Upsert(ctx, ns, &Edge{SourceURN: us, TargetURN: ent.URN, Type: "lineage", Properties: map[string]interface{}{"root": ent.URN}})
		}
		for _, ds := range downstreams {
			_ = s.edges.Upsert(ctx, ns, &Edge{SourceURN: ent.URN, TargetURN: ds, Type: "lineage", Properties: map[string]interface{}{"root": ent.URN}})
		}
	}

	return id, nil
}

func (s *Service) GetByURN(ctx context.Context, ns *namespace.Namespace, urn string) (Entity, error) {
	return s.repo.GetByURN(ctx, ns, urn)
}

func (s *Service) GetByID(ctx context.Context, id string) (Entity, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, ns *namespace.Namespace, flt Filter) ([]Entity, int, error) {
	entities, err := s.repo.GetAll(ctx, ns, flt)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.repo.GetCount(ctx, ns, flt)
	if err != nil {
		return entities, 0, nil
	}
	return entities, count, nil
}

func (s *Service) GetTypes(ctx context.Context, ns *namespace.Namespace) (map[Type]int, error) {
	return s.repo.GetTypes(ctx, ns)
}

func (s *Service) Delete(ctx context.Context, ns *namespace.Namespace, urn string) error {
	if err := s.repo.Delete(ctx, ns, urn); err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	if s.edges != nil {
		_ = s.edges.DeleteByURN(ctx, ns, urn)
	}
	return nil
}

func (s *Service) Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	if s.hybrid != nil && (cfg.Mode == SearchModeSemantic || cfg.Mode == SearchModeHybrid) {
		return s.hybrid.Search(ctx, cfg)
	}
	if s.search != nil {
		return s.search.Search(ctx, cfg)
	}
	return nil, nil
}

func (s *Service) Suggest(ctx context.Context, ns *namespace.Namespace, text string, limit int) ([]string, error) {
	if s.search != nil {
		return s.search.Suggest(ctx, ns, text, limit)
	}
	return nil, nil
}

// GetContext assembles a context subgraph around an entity.
func (s *Service) GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*ContextGraph, error) {
	ent, err := s.repo.GetByURN(ctx, ns, urn)
	if err != nil {
		return nil, fmt.Errorf("get entity: %w", err)
	}

	cg := &ContextGraph{Entity: ent}

	if s.edges != nil {
		_ = depth // TODO: use depth for multi-hop traversal
		outgoing, _ := s.edges.GetBySource(ctx, ns, urn, EdgeFilter{Current: true})
		incoming, _ := s.edges.GetByTarget(ctx, ns, urn, EdgeFilter{Current: true})
		cg.Edges = append(outgoing, incoming...)

		seen := map[string]bool{urn: true}
		for _, e := range cg.Edges {
			relURN := e.TargetURN
			if relURN == urn {
				relURN = e.SourceURN
			}
			if seen[relURN] {
				continue
			}
			seen[relURN] = true
			if rel, err := s.repo.GetByURN(ctx, ns, relURN); err == nil {
				cg.Related = append(cg.Related, rel)
			}
		}
	}

	return cg, nil
}

// GetImpact returns downstream entities affected by changes to the given entity.
func (s *Service) GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]Edge, error) {
	if s.edges == nil {
		return nil, nil
	}
	if depth <= 0 {
		depth = 3
	}
	return s.edges.GetDownstream(ctx, ns, urn, depth)
}

// ContextGraph is the assembled context subgraph for an entity.
type ContextGraph struct {
	Entity  Entity   `json:"entity"`
	Edges   []Edge   `json:"edges,omitempty"`
	Related []Entity `json:"related,omitempty"`
}
