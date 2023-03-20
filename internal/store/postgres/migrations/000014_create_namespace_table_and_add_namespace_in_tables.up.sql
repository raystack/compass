BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS namespaces (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  name text UNIQUE NOT NULL,
  state text,
  metadata jsonb,
  created_at timestamp DEFAULT NOW(),
  updated_at timestamp DEFAULT NOW(),
  deleted_at timestamp
);

ALTER TABLE assets ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_assets_namespace_id ON assets(namespace_id);

ALTER TABLE assets_versions ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_assets_versions_namespace_id ON assets_versions(namespace_id);

ALTER TABLE asset_owners ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_asset_owners_namespace_id ON asset_owners(namespace_id);

ALTER TABLE asset_probes ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_asset_probes_namespace_id ON asset_probes(namespace_id);

ALTER TABLE tags ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_tags_namespace_id ON tags(namespace_id);

ALTER TABLE tag_templates ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_tag_templates_namespace_id ON tag_templates(namespace_id);

ALTER TABLE tag_template_fields ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_tag_template_fields_namespace_id ON tag_template_fields(namespace_id);

ALTER TABLE users ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_users_namespace_id ON users(namespace_id);

ALTER TABLE stars ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_stars_namespace_id ON stars(namespace_id);

ALTER TABLE lineage_graph ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_lineage_graph_namespace_id ON lineage_graph(namespace_id);

ALTER TABLE discussions ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_discussions_namespace_id ON discussions(namespace_id);

ALTER TABLE comments ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_comments_namespace_id ON comments(namespace_id);

-- include namespace in index
ALTER TABLE asset_probes DROP CONSTRAINT IF EXISTS asset_probes_asset_urn_fkey;
DROP INDEX IF EXISTS assets_idx_urn;
CREATE UNIQUE INDEX assets_idx_urn ON assets(namespace_id,urn);
ALTER TABLE asset_probes ADD CONSTRAINT asset_probes_asset_urn_fkey FOREIGN KEY (namespace_id, asset_urn)
    REFERENCES assets(namespace_id, urn) ON DELETE CASCADE ON UPDATE CASCADE;

DROP INDEX IF EXISTS assets_versions_idx_urn_type_service_version;
CREATE UNIQUE INDEX assets_versions_idx_urn_version ON assets_versions(namespace_id,urn,version);

DROP INDEX IF EXISTS users_idx_uuid;
CREATE UNIQUE INDEX users_idx_uuid ON users(namespace_id,uuid);

DROP INDEX IF EXISTS users_idx_email;
CREATE UNIQUE INDEX users_idx_email ON users(namespace_id,email);

COMMIT;