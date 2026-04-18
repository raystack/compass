package mcp

import (
	"context"
	"fmt"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleGetDocuments(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if s.documentService == nil {
		return gomcp.NewToolResultError("document service not configured"), nil
	}

	urn := gomcp.ParseString(req, "urn", "")
	if urn == "" {
		return gomcp.NewToolResultError("'urn' parameter is required"), nil
	}

	docs, err := s.documentService.GetByEntityURN(ctx, getNamespace(ctx), urn)
	if err != nil {
		return gomcp.NewToolResultError("get documents failed: " + err.Error()), nil
	}

	if len(docs) == 0 {
		return gomcp.NewToolResultText(fmt.Sprintf("No documents found for %s.", urn)), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d documents for %s:\n\n", len(docs), urn)
	for _, d := range docs {
		fmt.Fprintf(&b, "### %s\n", d.Title)
		if d.Source != "" {
			fmt.Fprintf(&b, "Source: %s", d.Source)
			if d.SourceID != "" {
				fmt.Fprintf(&b, " (ID: %s)", d.SourceID)
			}
			fmt.Fprintf(&b, "\n")
		}
		body := d.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		fmt.Fprintf(&b, "%s\n\n", body)
	}
	return gomcp.NewToolResultText(b.String()), nil
}
