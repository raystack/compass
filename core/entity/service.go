package entity

import (
	"context"
	"fmt"
	"sort"

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
	docs     DocumentFetcher
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

// WithDocumentFetcher enables document fetching for context assembly.
func (s *Service) WithDocumentFetcher(d DocumentFetcher) {
	s.docs = d
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

// maxContextDepth caps the maximum traversal depth for context queries.
const maxContextDepth = 5

// GetContext assembles a context subgraph around an entity.
func (s *Service) GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*ContextGraph, error) {
	ent, err := s.repo.GetByURN(ctx, ns, urn)
	if err != nil {
		return nil, fmt.Errorf("get entity: %w", err)
	}

	cg := &ContextGraph{Entity: ent}

	if s.edges != nil {
		if depth <= 0 {
			depth = 1
		}
		if depth > maxContextDepth {
			depth = maxContextDepth
		}

		cg.Edges, err = s.edges.GetBidirectional(ctx, ns, urn, depth)
		if err != nil {
			return nil, fmt.Errorf("get context edges: %w", err)
		}

		seen := map[string]bool{urn: true}
		for _, e := range cg.Edges {
			for _, candidate := range []string{e.SourceURN, e.TargetURN} {
				if seen[candidate] {
					continue
				}
				seen[candidate] = true
				if rel, err := s.repo.GetByURN(ctx, ns, candidate); err == nil {
					cg.Related = append(cg.Related, rel)
				}
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

// AssembleContext intelligently assembles a context window for AI agent tasks.
func (s *Service) AssembleContext(ctx context.Context, ns *namespace.Namespace, req AssemblyRequest) (*AssembledContext, error) {
	req = validateAssemblyRequest(req)

	// Phase 1: Seed Resolution
	var seeds []Entity
	if len(req.SeedURNs) > 0 {
		for _, urn := range req.SeedURNs {
			e, err := s.repo.GetByURN(ctx, ns, urn)
			if err != nil {
				continue // skip missing seeds
			}
			seeds = append(seeds, e)
		}
	} else {
		results, err := s.Search(ctx, SearchConfig{
			Text:       req.Query,
			MaxResults: 3,
			Namespace:  ns,
		})
		if err == nil {
			for _, r := range results {
				seeds = append(seeds, Entity{
					URN: r.URN, Type: Type(r.Type), Name: r.Name,
					Description: r.Description, Source: r.Source,
				})
			}
		}
	}

	if len(seeds) == 0 {
		return &AssembledContext{Query: req.Query, Intent: req.Intent, TokenBudget: req.TokenBudget}, nil
	}

	// Phase 2: Graph Expansion
	var allEdges []Edge
	relatedURNs := map[string]int{} // URN -> min distance from any seed
	seedURNSet := map[string]bool{}
	for _, seed := range seeds {
		seedURNSet[seed.URN] = true
	}

	if s.edges != nil {
		for _, seed := range seeds {
			edges, err := s.edges.GetBidirectional(ctx, ns, seed.URN, req.Depth)
			if err != nil {
				continue
			}
			for _, e := range edges {
				allEdges = append(allEdges, e)
				otherURN := e.TargetURN
				if e.TargetURN == seed.URN {
					otherURN = e.SourceURN
				}
				if _, exists := relatedURNs[otherURN]; !exists {
					relatedURNs[otherURN] = 1 // simplified distance
				}
			}
		}
	}

	// Fetch related entities by URN (skip seeds, already have them)
	var relatedEntities []ScoredEntity
	for urn, dist := range relatedURNs {
		if seedURNSet[urn] {
			continue
		}
		e, err := s.repo.GetByURN(ctx, ns, urn)
		if err != nil {
			continue
		}
		score := 1.0 / float64(dist+1)
		score *= intentWeight(req.Intent, e.Type, dist)
		relatedEntities = append(relatedEntities, ScoredEntity{Entity: e, Score: score, Distance: dist})
	}

	// Sort by score descending
	sort.Slice(relatedEntities, func(i, j int) bool {
		return relatedEntities[i].Score > relatedEntities[j].Score
	})

	// Phase 3: Document Fetching
	var docs []FetchedDocument
	if s.docs != nil {
		for _, seed := range seeds {
			fetched, err := s.docs.GetDocumentsForEntity(ctx, ns, seed.URN)
			if err == nil {
				docs = append(docs, fetched...)
			}
		}
		limit := 7
		if limit > len(relatedEntities) {
			limit = len(relatedEntities)
		}
		for i := 0; i < limit; i++ {
			fetched, err := s.docs.GetDocumentsForEntity(ctx, ns, relatedEntities[i].Entity.URN)
			if err == nil {
				docs = append(docs, fetched...)
			}
		}
	}

	// Phase 4+5: Budget Packing
	result := &AssembledContext{
		Query:       req.Query,
		Intent:      req.Intent,
		TokenBudget: req.TokenBudget,
		Stats: AssemblyStats{
			EntitiesConsidered: len(seeds) + len(relatedEntities),
			GraphDepth:         req.Depth,
		},
	}

	budget := req.TokenBudget

	// Seeds always included
	for _, seed := range seeds {
		tokens := estimateEntityTokens(seed)
		budget -= tokens
		result.Seeds = append(result.Seeds, seed)
		result.Entities = append(result.Entities, ScoredEntity{Entity: seed, Score: 1.0, Distance: 0})
	}

	// Pack related entities
	for _, se := range relatedEntities {
		tokens := estimateEntityTokens(se.Entity)
		if budget-tokens < 0 {
			result.Truncated = true
			continue
		}
		budget -= tokens
		result.Entities = append(result.Entities, se)
		result.Stats.EntitiesIncluded++
	}

	// Pack edges (compact, low token cost)
	dedupEdges := deduplicateEdges(allEdges)
	for _, edge := range dedupEdges {
		tokens := estimateTokens(edge.SourceURN + edge.TargetURN + edge.Type)
		if budget-tokens < 0 {
			result.Truncated = true
			break
		}
		budget -= tokens
		result.Edges = append(result.Edges, edge)
	}

	// Pack documents (largest items, pack what fits)
	for _, doc := range docs {
		tokens := estimateDocTokens(doc)
		if budget-tokens < 0 {
			result.Truncated = true
			continue
		}
		budget -= tokens
		result.Documents = append(result.Documents, doc)
		result.Stats.DocumentsFetched++
	}

	result.TokensUsed = req.TokenBudget - budget
	result.Stats.EntitiesIncluded += len(seeds)

	return result, nil
}
