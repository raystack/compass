BEGIN;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS namespaces (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  name text UNIQUE,
  state text,
  metadata jsonb,
  created_at timestamp DEFAULT NOW(),
  updated_at timestamp DEFAULT NOW(),
  deleted_at timestamp
);
ALTER TABLE assets ADD COLUMN namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX idx_assets_namespace_id ON assets(namespace_id);
COMMIT;