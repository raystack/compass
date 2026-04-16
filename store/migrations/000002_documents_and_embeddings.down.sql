-- Drop documents table
DROP TABLE IF EXISTS documents CASCADE;

-- Revert embeddings → chunks
ALTER TABLE embeddings DROP COLUMN IF EXISTS content_type;
ALTER TABLE embeddings DROP COLUMN IF EXISTS content_id;

-- Revert vector dimension from 768 back to 1536
DROP INDEX IF EXISTS idx_embeddings_vector;
ALTER TABLE embeddings ALTER COLUMN embedding TYPE vector(1536);
CREATE INDEX chunks_embedding_idx ON embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

DROP POLICY IF EXISTS embeddings_ns ON embeddings;
CREATE POLICY chunks_ns ON embeddings
    USING (namespace_id = current_setting('app.current_tenant')::uuid);

ALTER INDEX IF EXISTS idx_embeddings_vector RENAME TO chunks_embedding_idx;
ALTER INDEX IF EXISTS idx_embeddings_entity_urn RENAME TO idx_chunks_entity_urn;
ALTER INDEX IF EXISTS idx_embeddings_namespace RENAME TO idx_chunks_namespace;
DROP INDEX IF EXISTS idx_embeddings_content_id;

ALTER TABLE embeddings RENAME TO chunks;
