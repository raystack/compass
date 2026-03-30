# Internals

This document details how Compass works under the hood. It covers the search architecture, storage internals, and multi-tenancy model.

## Search Architecture

All search in Compass is Postgres-native, combining keyword, fuzzy, and semantic strategies with no external search engine dependencies.

### Postgres-Native Search

#### Full-Text Search (tsvector)

Entities are indexed using PostgreSQL's built-in full-text search. A `search_vector` generated column is maintained on the entities table with weighted fields:

- **Weight A:** URN and name (highest relevance)
- **Weight B:** Description
- **Weight C:** Source and service metadata

GIN indexes on the search vector enable fast full-text queries.

#### Fuzzy Matching (pg_trgm)

Trigram indexes powered by the `pg_trgm` extension support typo-tolerant and partial matching. This handles cases where users misspell entity names or search with partial terms.

#### Semantic Search (pgvector)

Vector embeddings are stored in a chunks table and indexed for cosine similarity search using pgvector. When an entity is created or updated, its semantic content (description, properties, labels) is embedded and stored. Semantic search finds conceptually related entities even when the exact terms don't overlap.

#### Hybrid Ranking

Results from keyword and semantic search are combined using Reciprocal Rank Fusion (RRF). This produces a single ranked list that balances keyword precision with semantic recall.

## Entity Storage

### Temporal Model

Entities in Compass are temporal. Each entity version carries `valid_from` and `valid_to` timestamps, allowing Compass to track how entities and their properties evolve over time. This supports queries like "what did this entity look like last week" and "what changed in the last 24 hours."

### Graph Edges

Relationships between entities are stored as typed, directed edges. Each edge has a type (lineage, ownership, documentation, etc.) and optional properties. Edges are also temporal, capturing when relationships were established and when they ended.

Graph traversal uses recursive Common Table Expressions (CTEs) in PostgreSQL, enabling multi-hop queries without external graph database dependencies.

## PostgreSQL Multi-Tenancy

To enforce multi-tenant restrictions at the database level, [Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) is used. RLS requires Postgres users used for application database connection not to be a table owner or a superuser, else all RLS policies are bypassed by default. That means the Postgres user that runs migrations and the user that serves the app should be different.

To create a postgres user:

```sql
CREATE USER "compass_user" WITH PASSWORD 'compass';
GRANT CONNECT ON DATABASE "compass" TO "compass_user";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "compass_user";
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO "compass_user";
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO "compass_user";

ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES
ON TABLES TO "compass_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT USAGE ON SEQUENCES TO "compass_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT EXECUTE ON FUNCTIONS TO "compass_user";
```

A middleware looks for `x-namespace` header to extract tenant id. If not found, it falls back to the `default` namespace. The same can be passed in a JWT token of Authentication Bearer with `namespace_id` as a claim.
