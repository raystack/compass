-- ============================================================
-- Rename chunks → embeddings
-- ============================================================

ALTER TABLE chunks RENAME TO embeddings;
ALTER INDEX chunks_embedding_idx RENAME TO idx_embeddings_vector;
ALTER INDEX idx_chunks_entity_urn RENAME TO idx_embeddings_entity_urn;
ALTER INDEX idx_chunks_namespace RENAME TO idx_embeddings_namespace;

-- Change vector dimension from 1536 to 768 (Ollama nomic-embed-text default)
-- Must drop and recreate the HNSW index since it depends on the column type
DROP INDEX IF EXISTS idx_embeddings_vector;
ALTER TABLE embeddings ALTER COLUMN embedding TYPE vector(768);
CREATE INDEX idx_embeddings_vector ON embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 32, ef_construction = 256);

-- Add content tracking columns
ALTER TABLE embeddings ADD COLUMN content_id uuid;
ALTER TABLE embeddings ADD COLUMN content_type text DEFAULT 'entity';

CREATE INDEX idx_embeddings_content_id ON embeddings(content_id) WHERE content_id IS NOT NULL;

-- Update RLS policy name
DROP POLICY IF EXISTS chunks_ns ON embeddings;
CREATE POLICY embeddings_ns ON embeddings
    USING (namespace_id = current_setting('app.current_tenant')::uuid);

-- ============================================================
-- Documents (knowledge layer — prose, runbooks, annotations)
-- ============================================================

CREATE TABLE documents (
    id            uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    namespace_id  uuid NOT NULL REFERENCES namespaces(id),
    entity_urn    text NOT NULL,
    title         text NOT NULL,
    body          text NOT NULL,
    format        text DEFAULT 'markdown',
    source        text,
    source_id     text,
    properties    jsonb DEFAULT '{}',
    created_at    timestamptz DEFAULT now(),
    updated_at    timestamptz DEFAULT now(),
    UNIQUE (namespace_id, entity_urn, source, source_id)
);

CREATE INDEX idx_documents_entity_urn ON documents(namespace_id, entity_urn);
CREATE INDEX idx_documents_source ON documents(source);

ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY documents_ns ON documents
    USING (namespace_id = current_setting('app.current_tenant')::uuid);
