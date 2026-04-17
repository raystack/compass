-- Compass v2: Fresh schema
-- All search is Postgres-native: tsvector + pg_trgm + pgvector.

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================================
-- Namespaces (multi-tenancy root)
-- ============================================================

CREATE TABLE namespaces (
    id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name        text NOT NULL UNIQUE,
    state       text,
    metadata    jsonb,
    created_at  timestamptz DEFAULT now(),
    updated_at  timestamptz DEFAULT now(),
    deleted_at  timestamptz
);

-- ============================================================
-- Users
-- ============================================================

CREATE TABLE users (
    id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    namespace_id uuid NOT NULL REFERENCES namespaces(id),
    uuid        text,
    email       text,
    provider    text,
    created_at  timestamptz DEFAULT now(),
    updated_at  timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_uuid ON users(uuid) WHERE uuid IS NOT NULL AND uuid != '';
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL AND email != '';
CREATE INDEX idx_users_namespace ON users(namespace_id);

-- ============================================================
-- Entities (core knowledge layer)
-- ============================================================

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
CREATE INDEX idx_entities_search ON entities USING GIN(search_vector);
CREATE INDEX idx_entities_name_trgm ON entities USING GIN(name gin_trgm_ops);
CREATE INDEX idx_entities_urn_trgm ON entities USING GIN(urn gin_trgm_ops);

-- ============================================================
-- Edges (typed, temporal relationships between entities)
-- ============================================================

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

-- ============================================================
-- Chunks (vector embeddings for semantic search)
-- ============================================================

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
    WITH (m = 32, ef_construction = 256);
CREATE INDEX idx_chunks_entity_urn ON chunks(entity_urn);
CREATE INDEX idx_chunks_namespace ON chunks(namespace_id);

-- ============================================================
-- Stars (users starring entities)
-- ============================================================

CREATE TABLE stars (
    id          bigserial PRIMARY KEY,
    namespace_id uuid NOT NULL REFERENCES namespaces(id),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_id   uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    created_at  timestamptz DEFAULT now(),
    updated_at  timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX idx_stars_user_entity ON stars(user_id, entity_id);
CREATE INDEX idx_stars_entity ON stars(entity_id);
CREATE INDEX idx_stars_namespace ON stars(namespace_id);

-- ============================================================
-- Row Level Security (all tables)
-- ============================================================

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE edges ENABLE ROW LEVEL SECURITY;
ALTER TABLE chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE stars ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_ns ON users USING (namespace_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY entities_ns ON entities USING (namespace_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY edges_ns ON edges USING (namespace_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY chunks_ns ON chunks USING (namespace_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY stars_ns ON stars USING (namespace_id = current_setting('app.current_tenant')::UUID);
