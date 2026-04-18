package mcp

import (
	"context"
	"fmt"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/entity"
)

func (s *Server) handleAssembleContext(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if s.entityService == nil {
		return gomcp.NewToolResultError("entity service not configured"), nil
	}

	query := gomcp.ParseString(req, "query", "")
	if query == "" {
		return gomcp.NewToolResultError("'query' parameter is required"), nil
	}

	var seedURNs []string
	if raw := gomcp.ParseString(req, "seed_urns", ""); raw != "" {
		for _, u := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(u); trimmed != "" {
				seedURNs = append(seedURNs, trimmed)
			}
		}
	}

	intent := entity.Intent(gomcp.ParseString(req, "intent", "general"))
	tokenBudget := gomcp.ParseInt(req, "token_budget", 4000)
	depth := gomcp.ParseInt(req, "depth", 2)

	assembled, err := s.entityService.AssembleContext(ctx, getNamespace(ctx), entity.AssemblyRequest{
		Query:       query,
		SeedURNs:    seedURNs,
		Intent:      intent,
		TokenBudget: tokenBudget,
		Depth:       depth,
	})
	if err != nil {
		return gomcp.NewToolResultError("assemble context failed: " + err.Error()), nil
	}

	return gomcp.NewToolResultText(formatAssembledContext(assembled)), nil
}

func formatAssembledContext(ac *entity.AssembledContext) string {
	var b strings.Builder

	truncated := "no"
	if ac.Truncated {
		truncated = "yes"
	}

	fmt.Fprintf(&b, "# Context: %s\n", ac.Query)
	fmt.Fprintf(&b, "Intent: %s | Tokens: %d/%d | Truncated: %s\n", ac.Intent, ac.TokensUsed, ac.TokenBudget, truncated)

	if len(ac.Seeds) > 0 {
		b.WriteString("\n## Seed Entities\n")
		for _, seed := range ac.Seeds {
			fmt.Fprintf(&b, "### %s (%s)\n", seed.Name, seed.Type)
			fmt.Fprintf(&b, "- URN: %s\n", seed.URN)
			if seed.Description != "" {
				fmt.Fprintf(&b, "- Description: %s\n", seed.Description)
			}
			if seed.Source != "" {
				fmt.Fprintf(&b, "- Source: %s\n", seed.Source)
			}
		}
	}

	// Related entities (skip seeds which have Distance=0)
	var related []entity.ScoredEntity
	for _, e := range ac.Entities {
		if e.Distance > 0 {
			related = append(related, e)
		}
	}
	if len(related) > 0 {
		b.WriteString("\n## Related Entities\n")
		for _, se := range related {
			fmt.Fprintf(&b, "### %s (%s) [relevance: %.2f, distance: %d]\n", se.Entity.Name, se.Entity.Type, se.Score, se.Distance)
			fmt.Fprintf(&b, "- URN: %s\n", se.Entity.URN)
			if se.Entity.Description != "" {
				fmt.Fprintf(&b, "- Description: %s\n", se.Entity.Description)
			}
		}
	}

	if len(ac.Edges) > 0 {
		b.WriteString("\n## Relationships\n")
		for _, edge := range ac.Edges {
			fmt.Fprintf(&b, "- %s --[%s]--> %s\n", edge.SourceURN, edge.Type, edge.TargetURN)
		}
	}

	if len(ac.Documents) > 0 {
		b.WriteString("\n## Documents\n")
		for _, doc := range ac.Documents {
			fmt.Fprintf(&b, "### %s (attached to %s)\n", doc.Title, doc.EntityURN)
			fmt.Fprintf(&b, "%s\n\n", doc.Body)
		}
	}

	fmt.Fprintf(&b, "\n---\nProvenance: %d entities considered, %d included, depth %d\n",
		ac.Stats.EntitiesConsidered, ac.Stats.EntitiesIncluded, ac.Stats.GraphDepth)

	return b.String()
}
