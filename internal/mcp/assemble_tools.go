package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func assembleContextTool() mcp.Tool {
	return mcp.NewTool("assemble_context",
		mcp.WithDescription("Assemble a curated context window for an AI agent task. Resolves seed entities (by URN or search), expands the knowledge graph, fetches related documents, and packs everything into a token-budgeted response optimized for the given intent."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("What you're trying to accomplish"),
		),
		mcp.WithString("seed_urns",
			mcp.Description("Comma-separated entity URNs to start from"),
		),
		mcp.WithString("intent",
			mcp.Description("Task intent: debug, build, analyze, govern, general"),
		),
		mcp.WithNumber("token_budget",
			mcp.Description("Max tokens in response (default: 4000)"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Graph traversal depth (default: 2, max: 5)"),
		),
	)
}
