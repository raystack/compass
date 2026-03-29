package mcp

import (
	"context"
	"net/http"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
)

// EntityService defines entity operations needed by the MCP server.
type EntityService interface {
	Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error)
	GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error)
	GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error)
}

// Server is the MCP server that exposes Compass as AI-agent tools.
type Server struct {
	entityService EntityService
	namespace     *namespace.Namespace
	mcpServer     *mcpserver.MCPServer
	httpServer    *mcpserver.StreamableHTTPServer
}

// New creates a new MCP server.
func New(entitySvc EntityService, ns *namespace.Namespace) *Server {
	s := &Server{
		entityService: entitySvc,
		namespace:     ns,
	}

	mcpSrv := mcpserver.NewMCPServer(
		"compass",
		"0.2.0",
		mcpserver.WithToolCapabilities(false),
	)

	mcpSrv.AddTool(searchEntitiesTool(), s.handleSearchEntities)
	mcpSrv.AddTool(getContextTool(), s.handleGetContext)
	mcpSrv.AddTool(impactAnalysisTool(), s.handleImpact)

	s.mcpServer = mcpSrv
	s.httpServer = mcpserver.NewStreamableHTTPServer(mcpSrv)

	return s
}

// Handler returns an http.Handler for mounting the MCP server.
func (s *Server) Handler() http.Handler {
	return s.httpServer
}
