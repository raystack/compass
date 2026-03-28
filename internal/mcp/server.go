package mcp

import (
	"context"
	"net/http"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/namespace"
)

// AssetService defines the asset operations needed by the MCP server.
type AssetService interface {
	SearchAssets(ctx context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error)
	GetAssetByID(ctx context.Context, id string) (asset.Asset, error)
	GetLineage(ctx context.Context, urn string, query asset.LineageQuery) (asset.Lineage, error)
	GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error)
	GetAllAssets(ctx context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error)
}

// Server is the MCP server that exposes Compass catalog as AI-agent tools.
type Server struct {
	assetService AssetService
	namespace    *namespace.Namespace
	mcpServer    *mcpserver.MCPServer
	httpServer   *mcpserver.StreamableHTTPServer
}

// New creates a new MCP server with the given dependencies.
func New(assetSvc AssetService, ns *namespace.Namespace) *Server {
	s := &Server{
		assetService: assetSvc,
		namespace:    ns,
	}

	mcpSrv := mcpserver.NewMCPServer(
		"compass",
		"0.1.0",
		mcpserver.WithToolCapabilities(false),
	)

	mcpSrv.AddTool(searchAssetsTool(), s.handleSearchAssets)
	mcpSrv.AddTool(getAssetTool(), s.handleGetAsset)
	mcpSrv.AddTool(getLineageTool(), s.handleGetLineage)
	mcpSrv.AddTool(listTypesTool(), s.handleListTypes)
	mcpSrv.AddTool(getAllAssetsTool(), s.handleGetAllAssets)

	s.mcpServer = mcpSrv
	s.httpServer = mcpserver.NewStreamableHTTPServer(mcpSrv)

	return s
}

// Handler returns an http.Handler for mounting the MCP server on an existing mux.
func (s *Server) Handler() http.Handler {
	return s.httpServer
}
