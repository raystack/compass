-- Entity Model v2: entities, edges, chunks (alongside existing assets tables)
-- All search is Postgres-native: tsvector for keyword, pg_trgm for fuzzy, pgvector for semantic.
-- No Elasticsearch dependency for the v2 entity system.

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Entities: the core knowledge layer
CREATE TABLE entities (
    id              uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    namespace_id    uuid NOT NULL REFERENCES namespaces(id),
    urn             text NOT NULL,
    type            text NOT NULL,
    name            text NOT NULL,
    description     text,
    properties      jsonb DEFAULT '{}',
    source          text,
    valid_from      timestamptz DEFAULT now(),
    valid_to        timestamptz,
    created_at      timestamptz DEFAULT now(),
    updated_at      timestamptz DEFAULT now(),

    -- Full-text search vector (auto-populated by trigger)
    search_vector   tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(urn, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(source, '')), 'C')
    ) STORED,

    UNIQUE (namespace_id, urn, valid_from)
);

CREATE INDEX idx_entities_ns_urn ON entities(namespace_id, urn);
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_current ON entities(valid_to) WHERE valid_to IS NULL;
CREATE INDEX idx_entities_properties ON entities USING GIN(properties);
CREATE INDEX idx_entities_source ON entities(source);

-- Full-text search index (GIN on tsvector)
CREATE INDEX idx_entities_search ON entities USING GIN(search_vector);

-- Trigram indexes for fuzzy matching on name and URN
CREATE INDEX idx_entities_name_trgm ON entities USING GIN(name gin_trgm_ops);
CREATE INDEX idx_entities_urn_trgm ON entities USING GIN(urn gin_trgm_ops);

ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
CREATE POLICY entities_ns_policy ON entities
    USING (namespace_id = current_setting('app.current_tenant')::UUID);

-- Edges: typed, temporal relationships between entities
CREATE TABLE edges (
    id              uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    namespace_id    uuid NOT NULL REFERENCES namespaces(id),
    source_urn      text NOT NULL,
    target_urn      text NOT NULL,
    type            text NOT NULL,
    properties      jsonb DEFAULT '{}',
    valid_from      timestamptz DEFAULT now(),
    valid_to        timestamptz,
    source          text,
    created_at      timestamptz DEFAULT now(),

    UNIQUE (namespace_id, source_urn, target_urn, type, valid_from)
);

CREATE INDEX idx_edges_source_urn ON edges(namespace_id, source_urn);
CREATE INDEX idx_edges_target_urn ON edges(namespace_id, target_urn);
CREATE INDEX idx_edges_type ON edges(type);
CREATE INDEX idx_edges_current ON edges(valid_to) WHERE valid_to IS NULL;

ALTER TABLE edges ENABLE ROW LEVEL SECURITY;
CREATE POLICY edges_ns_policy ON edges
    USING (namespace_id = current_setting('app.current_tenant')::UUID);

-- Chunks: vector embeddings for semantic search
CREATE TABLE chunks (
    id              uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    namespace_id    uuid NOT NULL REFERENCES namespaces(id),
    entity_urn      text NOT NULL,
    content         text NOT NULL,
    context         text,
    embedding       vector(1536) NOT NULL,
    position        int,
    heading         text,
    token_count     int,
    created_at      timestamptz DEFAULT now()
);

CREATE INDEX chunks_embedding_idx ON chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
CREATE INDEX idx_chunks_entity_urn ON chunks(entity_urn);
CREATE INDEX idx_chunks_namespace ON chunks(namespace_id);

ALTER TABLE chunks ENABLE ROW LEVEL SECURITY;
CREATE POLICY chunks_ns_policy ON chunks
    USING (namespace_id = current_setting('app.current_tenant')::UUID);
