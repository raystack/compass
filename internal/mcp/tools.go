package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func searchAssetsTool() mcp.Tool {
	return mcp.NewTool("search_assets",
		mcp.WithDescription("Search for data assets in the Compass catalog. Returns matching tables, topics, dashboards, and other assets."),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Search query text"),
		),
		mcp.WithString("types",
			mcp.Description("Comma-separated asset types to filter by (e.g. table,topic,dashboard)"),
		),
		mcp.WithString("services",
			mcp.Description("Comma-separated services to filter by (e.g. bigquery,kafka)"),
		),
		mcp.WithNumber("size",
			mcp.Description("Maximum number of results to return (default: 10)"),
		),
	)
}

func getAssetTool() mcp.Tool {
	return mcp.NewTool("get_asset",
		mcp.WithDescription("Get full details of a data asset by its ID (UUID) or URN. Returns schema, owners, description, labels, and metadata."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Asset ID (UUID) or URN"),
		),
	)
}

func getLineageTool() mcp.Tool {
	return mcp.NewTool("get_lineage",
		mcp.WithDescription("Get the lineage graph for a data asset. Shows upstream sources and downstream consumers."),
		mcp.WithString("urn",
			mcp.Required(),
			mcp.Description("URN of the asset to get lineage for"),
		),
		mcp.WithString("direction",
			mcp.Description("Lineage direction: upstream, downstream, or both (default: both)"),
		),
		mcp.WithNumber("level",
			mcp.Description("Number of hops to traverse (default: 1)"),
		),
	)
}

func listTypesTool() mcp.Tool {
	return mcp.NewTool("list_types",
		mcp.WithDescription("List all asset types in the catalog with their counts."),
	)
}

func getAllAssetsTool() mcp.Tool {
	return mcp.NewTool("get_all_assets",
		mcp.WithDescription("Browse and filter data assets in the catalog with pagination."),
		mcp.WithString("types",
			mcp.Description("Comma-separated asset types to filter by (e.g. table,topic)"),
		),
		mcp.WithString("services",
			mcp.Description("Comma-separated services to filter by (e.g. bigquery,kafka)"),
		),
		mcp.WithString("q",
			mcp.Description("Query string to filter assets by name"),
		),
		mcp.WithNumber("size",
			mcp.Description("Number of results per page (default: 20)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Offset for pagination (default: 0)"),
		),
	)
}
