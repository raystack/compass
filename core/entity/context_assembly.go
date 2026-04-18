package entity

import (
	"context"

	"github.com/raystack/compass/core/namespace"
)

// Intent describes the purpose of the context assembly.
type Intent string

const (
	IntentDebug   Intent = "debug"
	IntentBuild   Intent = "build"
	IntentAnalyze Intent = "analyze"
	IntentGovern  Intent = "govern"
	IntentGeneral Intent = "general"
)

// AssemblyRequest describes what context to assemble.
type AssemblyRequest struct {
	Query       string
	SeedURNs    []string
	Intent      Intent
	TokenBudget int
	Depth       int
}

// AssembledContext is the result of context assembly.
type AssembledContext struct {
	Query       string
	Intent      Intent
	Seeds       []Entity
	Entities    []ScoredEntity
	Edges       []Edge
	Documents   []FetchedDocument
	TokensUsed  int
	TokenBudget int
	Truncated   bool
	Stats       AssemblyStats
}

// ScoredEntity is an entity with a relevance score and graph distance.
type ScoredEntity struct {
	Entity   Entity
	Score    float64
	Distance int // graph distance from nearest seed (0 = seed itself)
}

// FetchedDocument is a document fetched for context assembly.
type FetchedDocument struct {
	Title     string
	Body      string
	Source    string
	SourceID  string
	EntityURN string
}

// AssemblyStats captures stats about the assembly process.
type AssemblyStats struct {
	EntitiesConsidered int
	EntitiesIncluded   int
	DocumentsFetched   int
	GraphDepth         int
}

// DocumentFetcher is implemented by document.Service via an adapter.
type DocumentFetcher interface {
	GetDocumentsForEntity(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]FetchedDocument, error)
}

// estimateTokens returns a rough token estimate for a string.
func estimateTokens(s string) int {
	return (len(s) + 3) / 4
}

// estimateEntityTokens estimates the token cost of an entity.
func estimateEntityTokens(e Entity) int {
	return estimateTokens(e.URN + e.Name + e.Description + e.Source + string(e.Type))
}

// estimateDocTokens estimates the token cost of a document.
func estimateDocTokens(d FetchedDocument) int {
	return estimateTokens(d.Title + d.Body + d.Source + d.SourceID + d.EntityURN)
}

// intentWeight returns a scoring multiplier based on intent, entity type, and distance.
func intentWeight(intent Intent, entityType Type, distance int) float64 {
	_ = distance // available for future use
	switch intent {
	case IntentDebug:
		// upstream entities and lineage edges weighted higher
		switch entityType {
		case TypeJob, TypeTopic, TypeTable:
			return 1.5
		}
	case IntentBuild:
		// downstream and docs weighted higher
		switch entityType {
		case TypeTable, TypeTopic, TypeApplication:
			return 1.5
		}
	case IntentAnalyze:
		// metrics, dashboards weighted higher
		switch entityType {
		case TypeMetric, TypeDashboard:
			return 1.5
		}
	case IntentGovern:
		// policies, ownership weighted higher
		switch entityType {
		case TypeApplication, TypeModel:
			return 1.5
		}
	}
	return 1.0
}

// validateAssemblyRequest applies defaults to the request.
func validateAssemblyRequest(req AssemblyRequest) AssemblyRequest {
	if req.TokenBudget <= 0 {
		req.TokenBudget = 4000
	}
	if req.Depth <= 0 {
		req.Depth = 2
	}
	if req.Depth > 5 {
		req.Depth = 5
	}
	if req.Intent == "" {
		req.Intent = IntentGeneral
	}
	return req
}

// deduplicateEdges removes duplicate edges based on source_urn+target_urn+type.
func deduplicateEdges(edges []Edge) []Edge {
	seen := make(map[string]bool)
	var result []Edge
	for _, e := range edges {
		key := e.SourceURN + "|" + e.TargetURN + "|" + e.Type
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, e)
	}
	return result
}
