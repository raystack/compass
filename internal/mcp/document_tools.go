package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func getDocumentsTool() mcp.Tool {
	return mcp.NewTool("get_documents",
		mcp.WithDescription("Get all documents (runbooks, annotations, decisions, knowledge) attached to an entity."),
		mcp.WithString("urn",
			mcp.Required(),
			mcp.Description("URN of the entity to get documents for"),
		),
	)
}
