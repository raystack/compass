# Roadmap

Compass was originally built as a search and discovery engine for data assets. It stored metadata, indexed it for text search, and let humans find what they needed. It tracked lineage, supported discussions and tagging, and served as the catalog layer in the Raystack ecosystem.

That was the right design for a world where humans were the primary consumers of metadata. The world has shifted. AI agents, copilots, and autonomous systems are now the fastest-growing consumers of organizational knowledge. They don't browse a catalog UI. They need structured context they can reason over, a graph they can traverse programmatically, and answers they can trust.

Compass is well positioned for this shift. It already stores the assets, the relationships, the lineage, the ownership. What changes is how that knowledge is indexed, queried, and served.

This document describes where Compass goes next.

## The Role

Meteor and Compass are two halves of one system. Meteor goes wide — connecting to every source, extracting rich metadata, and delivering it. Compass goes deep — resolving entities, storing the graph, making it queryable, and serving it to every consumer.

Meteor is the agent that roams your infrastructure. Compass is the brain that remembers what the agent found, figures out what's the same thing across sources, and answers questions about it.

As Meteor evolves from a flat metadata collector into a richer extraction and delivery pipeline, Compass evolves from a search catalog into the **graph construction, persistence, query, and serving layer**. The division of responsibility:

- **Meteor** owns collection, enrichment, and delivery of raw observations from sources.
- **Compass** owns entity resolution, graph construction, storage, indexing, querying, and serving.

Meteor sends raw observations — "I found this thing at this source with these properties." Compass resolves those observations against what it already knows, deduplicates, constructs the graph, and serves it. This keeps Meteor stateless and simple, while Compass — which already has the full graph context needed for resolution — owns the intelligence.

The consumer base expands from humans using a UI to AI agents calling APIs. Both remain first-class. But the AI serving path is where most of the new investment goes.

## What Changes

### From Text Search to Semantic and Hybrid Search

Compass search was originally keyword-based. You typed a term, the search engine returned matches ranked by text relevance. This works when you know what you're looking for and can name it.

AI agents and natural language queries work differently. A user asking "find me everything related to revenue" expects results even when the word "revenue" doesn't appear in the asset name or description. The table might be called `monthly_mrr`. The column might be `net_arr`. The business context lives in descriptions, labels, discussions, and relationships — not always in the exact term being searched.

The search layer needs to become hybrid:

- **Vector embeddings alongside the text index.** When Meteor pushes an asset to Compass, Compass indexes both the text fields and a vector embedding of the asset's semantic content — description, column names, labels, business glossary terms. pgvector supports dense vector search natively in PostgreSQL.
- **Hybrid ranking.** Combine keyword precision (exact matches on table names, column names, URNs) with semantic similarity (conceptual matches on meaning). A query for "customer churn" should surface a table called `user_retention_weekly` even though the words don't overlap.
- **Natural language query decomposition.** Accept free-form questions and decompose them into structured graph queries plus semantic search. "What tables does the revenue dashboard depend on?" is a lineage traversal. "Find something related to customer lifetime value" is a semantic search. Compass should handle both through the same interface.

### From Catalog with Lineage to Graph-Native Store

Compass stores lineage as directed edges between assets. That's a relational approximation of a graph. It supports single-level upstream/downstream traversal. For the AI era, the graph needs to be richer and more traversable.

**Richer relationship types.** Lineage is one kind of relationship. There are many others that matter for AI reasoning:

- `owns` — who is responsible for this asset
- `reads` / `writes` — which jobs or services interact with this asset
- `documented_by` — which wiki page or design doc explains this asset
- `tested_by` — which data quality checks cover this asset
- `derived_from` — semantic derivation, not just pipeline lineage
- `similar_to` — assets with overlapping schemas or semantics

Each edge type carries its own semantics and should be queryable independently. "Show me all assets owned by the payments team" and "show me all assets downstream of this Kafka topic" are both graph queries over different edge types.

**Multi-hop traversal.** The current API supports lineage queries with a depth parameter, but the query model is limited to a single direction from a single starting node. Real questions require richer traversal:

- "If I change this table's schema, which dashboards break?" — traverse downstream across multiple asset types.
- "This metric dropped. What changed upstream in the last 24 hours?" — traverse upstream with a temporal filter.
- "Show me the full data flow from this Kafka topic to the executive dashboard." — path query across arbitrary depth.

The graph query API should support expressive traversal patterns, not just single-node neighborhood lookups.

**Entity resolution and identity.** The same logical thing appears across multiple sources — a BigQuery table, a dbt model, a Tableau datasource, and an Airflow task may all reference the same entity. Compass owns entity resolution. When Meteor delivers raw observations from different sources, Compass matches them against its existing graph, recognizes duplicates, merges facets, and maintains a unified entity identity. One entity, multiple source representations, queryable as either the unified entity or any individual facet. Compass is the right place for this because it has the full graph context needed to make resolution decisions — it can compare incoming observations against everything it already knows.

**Graph-aware ranking.** Assets that are highly connected — many dependents, many consumers, central in the lineage graph — are more important than orphaned tables nobody uses. Search ranking should factor in graph centrality and connectivity, not just text relevance.

### From Human UI to AI Serving Layer

This is the most important shift. Compass needs to serve AI agents as a primary consumer class, not just humans.

**MCP server.** Compass should expose itself as an MCP (Model Context Protocol) server. Any MCP-compatible AI system — Claude, coding assistants, custom agents — can connect to Compass and get tools like:

- `search_assets` — semantic search across the catalog
- `get_lineage` — traverse the graph from any starting node
- `get_schema` — full schema with column descriptions and types
- `get_owners` — who to ask about this asset
- `get_context` — a composed context document combining schema, lineage, ownership, recent changes, and quality signals into one coherent response

This is the fastest path to making Compass useful for AI. It doesn't require changing the storage layer — it's a new serving interface over existing data.

**Context composition.** AI agents have limited context windows. Compass should be smart about what it returns. Instead of raw JSON payloads, Compass should produce **context documents** — structured summaries that combine the most relevant information about an asset or a set of assets into a format an LLM can reason over efficiently.

A context document for a table might include: schema with column descriptions, immediate upstream and downstream lineage, current owners, last schema change, known data quality issues from discussions, freshness score. All in one response, formatted for LLM consumption, sized appropriately for the context window.

**Configurable depth and detail.** An agent exploring broadly needs summaries. An agent generating SQL needs full column-level detail with types and constraints. An agent investigating a data quality issue needs lineage and change history. Compass should serve all three use cases through the same API with different detail levels.

**Tool definition generation.** Compass knows the schema of every table, the parameters of every API, the structure of every dataset. It can auto-generate tool/function definitions that let AI agents interact with those assets directly — complete with column names, types, and descriptions as parameter documentation.

### From Passive Store to Active Knowledge

Compass today is passive. Data goes in, queries come out. For AI agents that need to stay current, Compass should be more active.

**Change feeds.** A streaming API of graph mutations. When a new asset appears, a schema changes, ownership transfers, or lineage shifts, downstream consumers should be notified. AI agents maintaining cached context can subscribe and keep their view fresh without polling.

**Freshness and quality signals.** Compass already supports probes — health and status metadata per asset. This should expand into a proper quality layer: freshness scores (how recently was this data updated), completeness metrics (are there null columns that shouldn't be null), anomaly flags (did the row count drop 90% today). When an AI agent picks a data source, it should know how trustworthy that source is right now, not just that it exists.

**Usage tracking.** Which assets are queried most frequently? Which are accessed by the most teams? Which have never been read? Usage data helps AI agents recommend relevant assets, prioritize popular sources, and identify candidates for deprecation.

**Subscriptions.** "Notify me when the schema of this table changes." "Alert when a new asset appears in this namespace." "Watch for lineage changes affecting this dashboard." AI agents monitoring data pipelines need reactive notifications, not just pull-based queries.

### Knowledge Layer Beyond Technical Metadata

Technical metadata — schemas, lineage, service information — is necessary but not sufficient. AI agents need business context to be genuinely useful.

**Business glossary.** First-class support for business terms, metric definitions, and domain concepts as entities in the graph, linked to the technical assets that implement them. "Revenue" as a concept connected to the three tables and two dashboards that compute it. "Churn" linked to the definition agreed upon by the analytics team, the SQL that calculates it, and the dashboard that displays it. When an AI agent needs to reason about a business concept, it should find both the definition and the implementation.

**Tribal knowledge capture.** Discussions and comments exist in Compass today, but they function as social features. They should be repositioned as knowledge capture. When someone explains in a discussion why a table has a non-obvious schema, documents a known data quality issue, or records the context behind a design decision — that's context an AI agent needs. These contributions should be surfaced alongside technical metadata, not buried in a comments tab.

**Documentation links.** Assets should link to external documentation — runbooks, design docs, ADRs, wiki pages. Not storing the full content, but maintaining the graph edge that says "this table is documented in this Confluence page" or "this pipeline's architecture is described in this ADR." Compass becomes the index that connects technical assets to human-written context wherever it lives.

## Architecture Direction

```
                     Meteor
                       │
                       │ raw observations + relationships
                       v
┌──────────────────────────────────────────────────┐
│                    Compass                        │
│                                                   │
│  ┌──────────────────┐                             │
│  │  Entity Resolver  │                             │
│  │  • deduplication  │                             │
│  │  • identity merge │                             │
│  └────────┬─────────┘                              │
│           v                                        │
│  ┌─────────────┐  ┌──────────────┐               │
│  │ Graph Store  │  │ Vector Index │               │
│  │ (Postgres)   │  │ (Semantic)   │               │
│  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                        │
│         v                 v                        │
│  ┌─────────────────────────────────┐              │
│  │         Query Engine            │              │
│  │  • graph traversal              │              │
│  │  • hybrid search (text+vector)  │              │
│  │  • context composition          │              │
│  │  • relevance + graph ranking    │              │
│  └───────────┬─────────────────────┘              │
│              │                                     │
│    ┌─────────┼──────────┬──────────┐              │
│    v         v          v          v              │
│ ┌──────┐ ┌────────┐ ┌────────┐ ┌────────┐       │
│ │ gRPC │ │  MCP   │ │Change  │ │  REST  │       │
│ │      │ │ Server │ │ Feed   │ │        │       │
│ └──┬───┘ └───┬────┘ └───┬────┘ └───┬────┘       │
│    │         │          │          │              │
└────┼─────────┼──────────┼──────────┼──────────────┘
     v         v          v          v
  Human UI  AI Agents  Downstream  Integrations
                       Systems
```

The key additions over today's architecture:

- **Entity resolver** that matches incoming observations from Meteor against the existing graph, deduplicates, and merges identities.
- **Vector index** with pgvector for semantic search.
- **Query engine** that can do graph traversal, text search, and semantic search in a single query.
- **MCP server** as a first-class API surface alongside gRPC and REST.
- **Change feed** for reactive downstream consumers including AI agents.

The existing gRPC/REST APIs and Postgres storage remain. Elasticsearch has been replaced by Postgres-native search (tsvector, pg_trgm, pgvector).

## Priorities

**First: MCP server.** Wrap the existing search, lineage, and asset APIs as MCP tools. This is a thin serving layer over what already exists and immediately makes Compass usable by any MCP-compatible AI system. No storage changes, no new indexing — just a new way to access what's already there.

**Second: Context composition.** Add an endpoint that takes an asset URN or a natural language query and returns a composed context document. Schema, lineage, ownership, descriptions, quality signals — all combined into LLM-ready output. This changes the consumer experience without changing the storage model.

**Third: Semantic search.** Add vector embeddings to the indexing pipeline. When an asset is upserted, Compass indexes both the text and a vector embedding. Search becomes hybrid. This requires adding embedding generation and a vector index but doesn't change the data model.

**Fourth: Rich graph model.** Expand relationship types beyond lineage. Add entity resolution to match and merge observations from Meteor into unified entities. Add multi-hop traversal queries. Implement graph-aware ranking. This requires schema evolution and new query capabilities — it's the deepest change.

**Fifth: Active knowledge layer.** Change feeds, subscriptions, freshness scoring, usage tracking, business glossary. This turns Compass from a store you query into a system that actively keeps consumers informed and provides business context alongside technical metadata.

## What Stays the Same

- **Postgres as the source of truth.** The transactional store is solid and stays. All search is now Postgres-native (tsvector, pg_trgm, pgvector) — no external search engine dependencies.
- **gRPC and REST APIs.** Existing integrations keep working. MCP is an additional serving layer.
- **Asset model.** The core asset structure — URN, type, service, schema, lineage, owners — remains the foundation. It gets extended, not replaced.
- **Social features.** Starring, discussions, tagging — these stay and become more valuable as knowledge capture mechanisms for AI consumption.
- **Namespace-based multi-tenancy.** The isolation model is sound and carries forward.

## The Bet

The metadata catalog market was about helping humans find data. That problem is largely solved. The next problem is helping AI systems understand an organization — its data, services, relationships, ownership, context, and meaning.

Compass already stores most of what AI needs. The gap is in how it's indexed (semantic, not just keyword), how it's queried (graph traversal, not just search), and how it's served (MCP and context documents, not just REST endpoints).

Meteor collects the knowledge. Compass makes it useful. Together they become the context layer that sits between an organization's infrastructure and its AI systems.
