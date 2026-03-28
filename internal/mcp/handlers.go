package mcp

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/raystack/compass/core/asset"
)

func (s *Server) handleSearchAssets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := mcp.ParseString(req, "text", "")
	if text == "" {
		return mcp.NewToolResultError("'text' parameter is required"), nil
	}

	size := mcp.ParseInt(req, "size", 10)

	cfg := asset.SearchConfig{
		Text:       strings.TrimSpace(text),
		MaxResults: size,
		Namespace:  s.namespace,
	}

	if types := mcp.ParseString(req, "types", ""); types != "" {
		cfg.Filters = map[string][]string{
			"type": strings.Split(types, ","),
		}
	}
	if services := mcp.ParseString(req, "services", ""); services != "" {
		if cfg.Filters == nil {
			cfg.Filters = make(map[string][]string)
		}
		cfg.Filters["service"] = strings.Split(services, ",")
	}

	results, err := s.assetService.SearchAssets(ctx, cfg)
	if err != nil {
		return mcp.NewToolResultError("failed to search assets: " + err.Error()), nil
	}

	return mcp.NewToolResultText(formatSearchResults(results)), nil
}

func (s *Server) handleGetAsset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := mcp.ParseString(req, "id", "")
	if id == "" {
		return mcp.NewToolResultError("'id' parameter is required"), nil
	}

	a, err := s.assetService.GetAssetByID(ctx, id)
	if err != nil {
		return mcp.NewToolResultError("failed to get asset: " + err.Error()), nil
	}

	return mcp.NewToolResultText(formatAsset(a)), nil
}

func (s *Server) handleGetLineage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	urn := mcp.ParseString(req, "urn", "")
	if urn == "" {
		return mcp.NewToolResultError("'urn' parameter is required"), nil
	}

	direction := asset.LineageDirection(mcp.ParseString(req, "direction", ""))
	if direction != "" && direction != asset.LineageDirectionUpstream && direction != asset.LineageDirectionDownstream {
		return mcp.NewToolResultError("'direction' must be 'upstream', 'downstream', or empty for both"), nil
	}

	level := mcp.ParseInt(req, "level", 1)

	lineage, err := s.assetService.GetLineage(ctx, urn, asset.LineageQuery{
		Level:          level,
		Direction:      direction,
		WithAttributes: true,
	})
	if err != nil {
		return mcp.NewToolResultError("failed to get lineage: " + err.Error()), nil
	}

	return mcp.NewToolResultText(formatLineage(urn, lineage)), nil
}

func (s *Server) handleListTypes(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flt, err := asset.NewFilterBuilder().Build()
	if err != nil {
		return mcp.NewToolResultError("failed to build filter: " + err.Error()), nil
	}

	types, err := s.assetService.GetTypes(ctx, flt)
	if err != nil {
		return mcp.NewToolResultError("failed to list types: " + err.Error()), nil
	}

	return mcp.NewToolResultText(formatTypes(types)), nil
}

func (s *Server) handleGetAllAssets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	size := mcp.ParseInt(req, "size", 20)
	offset := mcp.ParseInt(req, "offset", 0)

	fb := asset.NewFilterBuilder().
		Size(size).
		Offset(offset)

	if types := mcp.ParseString(req, "types", ""); types != "" {
		fb = fb.Types(types)
	}
	if services := mcp.ParseString(req, "services", ""); services != "" {
		fb = fb.Services(services)
	}
	if q := mcp.ParseString(req, "q", ""); q != "" {
		fb = fb.Q(q)
	}

	flt, err := fb.Build()
	if err != nil {
		return mcp.NewToolResultError("invalid filter: " + err.Error()), nil
	}

	assets, total, err := s.assetService.GetAllAssets(ctx, flt, true)
	if err != nil {
		return mcp.NewToolResultError("failed to get assets: " + err.Error()), nil
	}

	return mcp.NewToolResultText(formatAssets(assets, total)), nil
}
