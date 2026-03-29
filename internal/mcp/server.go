package mcp

import (
	"context"
	"net/http"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/entity"
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

// EntityServiceV2 defines the v2 entity operations for MCP tools.
type EntityServiceV2 interface {
	Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error)
	GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error)
	GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error)
}

// Server is the MCP server that exposes Compass catalog as AI-agent tools.
type Server struct {
	assetService  AssetService
	entityService EntityServiceV2
	namespace     *namespace.Namespace
	mcpServer     *mcpserver.MCPServer
	httpServer    *mcpserver.StreamableHTTPServer
}

// Option configures the MCP server.
type Option func(*Server)

// WithEntityService adds the v2 entity service for new MCP tools.
func WithEntityService(svc EntityServiceV2) Option {
	return func(s *Server) { s.entityService = svc }
}

// New creates a new MCP server with the given dependencies.
func New(assetSvc AssetService, ns *namespace.Namespace, opts ...Option) *Server {
	s := &Server{
		assetService: assetSvc,
		namespace:    ns,
	}
	for _, opt := range opts {
		opt(s)
	}

	mcpSrv := mcpserver.NewMCPServer(
		"compass",
		"0.2.0",
		mcpserver.WithToolCapabilities(false),
	)

	// v1 asset tools (retained)
	mcpSrv.AddTool(searchAssetsTool(), s.handleSearchAssets)
	mcpSrv.AddTool(getAssetTool(), s.handleGetAsset)
	mcpSrv.AddTool(getLineageTool(), s.handleGetLineage)
	mcpSrv.AddTool(listTypesTool(), s.handleListTypes)
	mcpSrv.AddTool(getAllAssetsTool(), s.handleGetAllAssets)

	// v2 entity tools
	mcpSrv.AddTool(searchEntitiesTool(), s.handleSearchEntities)
	mcpSrv.AddTool(getContextTool(), s.handleGetContext)
	mcpSrv.AddTool(impactAnalysisTool(), s.handleImpact)

	s.mcpServer = mcpSrv
	s.httpServer = mcpserver.NewStreamableHTTPServer(mcpSrv)

	return s
}

// Handler returns an http.Handler for mounting the MCP server on an existing mux.
func (s *Server) Handler() http.Handler {
	return s.httpServer
}
