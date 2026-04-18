package mcp

import (
	"context"
	"fmt"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/entity"
)

func (s *Server) handleSearchEntities(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if s.entityService == nil {
		return gomcp.NewToolResultError("entity service not configured"), nil
	}

	text := gomcp.ParseString(req, "text", "")
	if text == "" {
		return gomcp.NewToolResultError("'text' parameter is required"), nil
	}

	size := gomcp.ParseInt(req, "size", 10)
	mode := entity.SearchMode(gomcp.ParseString(req, "mode", "keyword"))

	cfg := entity.SearchConfig{
		Text:       strings.TrimSpace(text),
		MaxResults: size,
		Mode:       mode,
		Namespace:  getNamespace(ctx),
	}

	if types := gomcp.ParseString(req, "types", ""); types != "" {
		cfg.Filters = map[string][]string{"type": strings.Split(types, ",")}
	}
	if source := gomcp.ParseString(req, "source", ""); source != "" {
		if cfg.Filters == nil {
			cfg.Filters = make(map[string][]string)
		}
		cfg.Filters["source"] = []string{source}
	}

	results, err := s.entityService.Search(ctx, cfg)
	if err != nil {
		return gomcp.NewToolResultError("search failed: " + err.Error()), nil
	}

	return gomcp.NewToolResultText(formatEntitySearchResults(results)), nil
}

func (s *Server) handleGetContext(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if s.entityService == nil {
		return gomcp.NewToolResultError("entity service not configured"), nil
	}

	urn := gomcp.ParseString(req, "urn", "")
	if urn == "" {
		return gomcp.NewToolResultError("'urn' parameter is required"), nil
	}

	depth := gomcp.ParseInt(req, "depth", 2)

	cg, err := s.entityService.GetContext(ctx, getNamespace(ctx), urn, depth)
	if err != nil {
		return gomcp.NewToolResultError("get context failed: " + err.Error()), nil
	}

	return gomcp.NewToolResultText(formatContextGraph(cg)), nil
}

func (s *Server) handleImpact(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if s.entityService == nil {
		return gomcp.NewToolResultError("entity service not configured"), nil
	}

	urn := gomcp.ParseString(req, "urn", "")
	if urn == "" {
		return gomcp.NewToolResultError("'urn' parameter is required"), nil
	}

	depth := gomcp.ParseInt(req, "depth", 3)

	edges, err := s.entityService.GetImpact(ctx, getNamespace(ctx), urn, depth)
	if err != nil {
		return gomcp.NewToolResultError("impact analysis failed: " + err.Error()), nil
	}

	return gomcp.NewToolResultText(formatImpactAnalysis(urn, edges)), nil
}

// Formatters

func formatEntitySearchResults(results []entity.SearchResult) string {
	if len(results) == 0 {
		return "No entities found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d entities:\n\n", len(results))
	for _, r := range results {
		fmt.Fprintf(&b, "- **%s** (%s) — source: %s, urn: %s\n", r.Name, r.Type, r.Source, r.URN)
		if r.Description != "" {
			desc := r.Description
			if len(desc) > 120 {
				desc = desc[:120] + "..."
			}
			fmt.Fprintf(&b, "  %s\n", desc)
		}
	}
	return b.String()
}

func formatContextGraph(cg *entity.ContextGraph) string {
	var b strings.Builder
	e := cg.Entity
	fmt.Fprintf(&b, "## %s (%s)\n", e.Name, e.Type)
	fmt.Fprintf(&b, "URN: %s | Source: %s\n", e.URN, e.Source)
	if e.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", e.Description)
	}

	if len(cg.Edges) > 0 {
		b.WriteString("\n### Relationships\n")
		for _, edge := range cg.Edges {
			if edge.SourceURN == e.URN {
				fmt.Fprintf(&b, "  —[%s]→ %s\n", edge.Type, edge.TargetURN)
			} else {
				fmt.Fprintf(&b, "  ←[%s]— %s\n", edge.Type, edge.SourceURN)
			}
		}
	}

	if len(cg.Related) > 0 {
		b.WriteString("\n### Related Entities\n")
		for _, r := range cg.Related {
			fmt.Fprintf(&b, "- **%s** (%s) — %s\n", r.Name, r.Type, r.URN)
		}
	}

	return b.String()
}

func formatImpactAnalysis(urn string, edges []entity.Edge) string {
	if len(edges) == 0 {
		return fmt.Sprintf("No downstream dependencies found for %s.", urn)
	}

	affected := make(map[string]bool)
	for _, e := range edges {
		if e.Source != "" && e.SourceURN != urn {
			affected[e.SourceURN] = true
		}
		if e.TargetURN != urn {
			affected[e.TargetURN] = true
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Impact Analysis for %s\n\n", urn)
	fmt.Fprintf(&b, "**%d entities affected** across %d edges:\n\n", len(affected), len(edges))
	for _, e := range edges {
		fmt.Fprintf(&b, "  %s → %s\n", e.SourceURN, e.TargetURN)
	}
	return b.String()
}
