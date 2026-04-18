package mcp

import (
	"context"
	"net/http"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/middleware"
)

// EntityService defines entity operations needed by the MCP server.
type EntityService interface {
	Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error)
	GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error)
	GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error)
	AssembleContext(ctx context.Context, ns *namespace.Namespace, req entity.AssemblyRequest) (*entity.AssembledContext, error)
}

// DocumentService defines document operations needed by the MCP server.
type DocumentService interface {
	GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]document.Document, error)
}

// Server is the MCP server that exposes Compass as AI-agent tools.
type Server struct {
	entityService   EntityService
	documentService DocumentService
	mcpServer       *mcpserver.MCPServer
	httpServer      *mcpserver.StreamableHTTPServer
}

// New creates a new MCP server.
func New(entitySvc EntityService, docSvc DocumentService) *Server {
	s := &Server{
		entityService:   entitySvc,
		documentService: docSvc,
	}

	mcpSrv := mcpserver.NewMCPServer(
		"compass",
		"0.3.0",
		mcpserver.WithToolCapabilities(false),
	)

	mcpSrv.AddTool(searchEntitiesTool(), s.handleSearchEntities)
	mcpSrv.AddTool(getContextTool(), s.handleGetContext)
	mcpSrv.AddTool(impactAnalysisTool(), s.handleImpact)
	mcpSrv.AddTool(getDocumentsTool(), s.handleGetDocuments)
	mcpSrv.AddTool(assembleContextTool(), s.handleAssembleContext)

	s.mcpServer = mcpSrv
	s.httpServer = mcpserver.NewStreamableHTTPServer(mcpSrv)

	return s
}

// Handler returns an http.Handler for mounting the MCP server.
func (s *Server) Handler() http.Handler {
	return s.httpServer
}

// getNamespace returns the namespace from context, falling back to DefaultNamespace.
func getNamespace(ctx context.Context) *namespace.Namespace {
	return middleware.FetchNamespaceFromContext(ctx)
}
