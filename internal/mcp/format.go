package mcp

import (
	"fmt"
	"strings"

	"github.com/raystack/compass/core/asset"
)

// formatAsset formats an asset as LLM-friendly markdown text.
func formatAsset(a asset.Asset) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## %s (%s)\n", a.Name, a.Type)
	fmt.Fprintf(&b, "Service: %s | URN: %s\n", a.Service, a.URN)

	if a.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", a.Description)
	}

	if len(a.Owners) > 0 {
		names := make([]string, 0, len(a.Owners))
		for _, o := range a.Owners {
			if o.Email != "" {
				names = append(names, o.Email)
			} else {
				names = append(names, o.UUID)
			}
		}
		fmt.Fprintf(&b, "Owners: %s\n", strings.Join(names, ", "))
	}

	if a.URL != "" {
		fmt.Fprintf(&b, "URL: %s\n", a.URL)
	}

	if len(a.Labels) > 0 {
		pairs := make([]string, 0, len(a.Labels))
		for k, v := range a.Labels {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(&b, "Labels: %s\n", strings.Join(pairs, ", "))
	}

	formatAssetData(&b, a.Data)

	return b.String()
}

// formatAssetData formats the Data map, extracting schema columns if present.
func formatAssetData(b *strings.Builder, data map[string]interface{}) {
	if data == nil {
		return
	}

	// Extract schema/columns if present (common in table/topic assets)
	if columns, ok := extractColumns(data); ok && len(columns) > 0 {
		fmt.Fprintf(b, "\nColumns (%d):\n", len(columns))
		for _, col := range columns {
			name, _ := col["name"].(string)
			dataType, _ := col["data_type"].(string)
			desc, _ := col["description"].(string)

			if desc != "" {
				fmt.Fprintf(b, "  - %s (%s): %s\n", name, dataType, desc)
			} else {
				fmt.Fprintf(b, "  - %s (%s)\n", name, dataType)
			}
		}
	}
}

// extractColumns tries to find column definitions in asset data.
func extractColumns(data map[string]interface{}) ([]map[string]interface{}, bool) {
	// Try common paths: data.columns, data.schema.columns
	if cols, ok := data["columns"]; ok {
		return toMapSlice(cols)
	}
	if schema, ok := data["schema"].(map[string]interface{}); ok {
		if cols, ok := schema["columns"]; ok {
			return toMapSlice(cols)
		}
	}
	return nil, false
}

func toMapSlice(v interface{}) ([]map[string]interface{}, bool) {
	slice, ok := v.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]map[string]interface{}, 0, len(slice))
	for _, item := range slice {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result, len(result) > 0
}

// formatSearchResult formats a search result as a compact line.
func formatSearchResult(sr asset.SearchResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "- **%s** (%s) — service: %s, urn: %s", sr.Title, sr.Type, sr.Service, sr.URN)
	if sr.Description != "" {
		desc := sr.Description
		if len(desc) > 120 {
			desc = desc[:120] + "..."
		}
		fmt.Fprintf(&b, "\n  %s", desc)
	}
	return b.String()
}

// formatSearchResults formats a list of search results.
func formatSearchResults(results []asset.SearchResult) string {
	if len(results) == 0 {
		return "No assets found."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d assets:\n\n", len(results))
	for _, sr := range results {
		b.WriteString(formatSearchResult(sr))
		b.WriteString("\n")
	}
	return b.String()
}

// formatLineage formats lineage data as readable text.
func formatLineage(urn string, lineage asset.Lineage) string {
	if len(lineage.Edges) == 0 {
		return fmt.Sprintf("No lineage found for %s.", urn)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Lineage for %s (%d edges):\n\n", urn, len(lineage.Edges))

	upstreams := make([]string, 0)
	downstreams := make([]string, 0)

	for _, edge := range lineage.Edges {
		if edge.Target == urn {
			upstreams = append(upstreams, edge.Source)
		} else if edge.Source == urn {
			downstreams = append(downstreams, edge.Target)
		} else {
			// Transitive edges
			fmt.Fprintf(&b, "  %s → %s\n", edge.Source, edge.Target)
		}
	}

	if len(upstreams) > 0 {
		b.WriteString("Upstream (sources):\n")
		for _, u := range upstreams {
			fmt.Fprintf(&b, "  ← %s\n", u)
		}
	}

	if len(downstreams) > 0 {
		b.WriteString("Downstream (consumers):\n")
		for _, d := range downstreams {
			fmt.Fprintf(&b, "  → %s\n", d)
		}
	}

	return b.String()
}

// formatTypes formats asset type counts.
func formatTypes(types map[asset.Type]int) string {
	if len(types) == 0 {
		return "No asset types found."
	}

	var b strings.Builder
	b.WriteString("Asset types:\n\n")
	for t, count := range types {
		fmt.Fprintf(&b, "- %s: %d assets\n", t, count)
	}
	return b.String()
}

// formatAssets formats a list of assets as a summary list.
func formatAssets(assets []asset.Asset, total uint32) string {
	if len(assets) == 0 {
		return "No assets found."
	}

	var b strings.Builder
	if total > 0 {
		fmt.Fprintf(&b, "Showing %d of %d assets:\n\n", len(assets), total)
	} else {
		fmt.Fprintf(&b, "Found %d assets:\n\n", len(assets))
	}

	for _, a := range assets {
		fmt.Fprintf(&b, "- **%s** (%s) — service: %s, urn: %s\n", a.Name, a.Type, a.Service, a.URN)
	}
	return b.String()
}
