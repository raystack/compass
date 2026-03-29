package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func searchEntitiesTool() mcp.Tool {
	return mcp.NewTool("search_entities",
		mcp.WithDescription("Search the entity knowledge graph. Supports keyword, semantic, and hybrid search modes. Finds tables, services, pipelines, people, and any other entity type."),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Search query text"),
		),
		mcp.WithString("types",
			mcp.Description("Comma-separated entity types to filter by (e.g. table,topic,dashboard,pipeline)"),
		),
		mcp.WithString("source",
			mcp.Description("Filter by source system (e.g. bigquery,kafka)"),
		),
		mcp.WithString("mode",
			mcp.Description("Search mode: keyword, semantic, or hybrid (default: keyword)"),
		),
		mcp.WithNumber("size",
			mcp.Description("Maximum number of results (default: 10)"),
		),
	)
}

func getContextTool() mcp.Tool {
	return mcp.NewTool("get_context",
		mcp.WithDescription("Get full context about an entity: the entity itself, its relationships (edges), and related entities. The primary tool for understanding an entity in context."),
		mcp.WithString("urn",
			mcp.Required(),
			mcp.Description("URN of the entity"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Relationship traversal depth (default: 2)"),
		),
	)
}

func impactAnalysisTool() mcp.Tool {
	return mcp.NewTool("impact",
		mcp.WithDescription("Analyze downstream blast radius: what entities are affected if this entity changes?"),
		mcp.WithString("urn",
			mcp.Required(),
			mcp.Description("URN of the entity to analyze"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Downstream traversal depth (default: 3)"),
		),
	)
}
