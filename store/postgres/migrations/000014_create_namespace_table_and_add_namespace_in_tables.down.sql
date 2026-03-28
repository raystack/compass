BEGIN;

-- foreign key is a pain
ALTER TABLE asset_probes DROP CONSTRAINT IF EXISTS asset_probes_asset_urn_fkey;

DROP INDEX IF EXISTS assets_idx_urn;
CREATE UNIQUE INDEX assets_idx_urn ON assets(urn);

DROP INDEX IF EXISTS idx_assets_namespace_id;
ALTER TABLE assets DROP COLUMN IF EXISTS namespace_id;

ALTER TABLE asset_probes ADD CONSTRAINT asset_probes_asset_urn_fkey FOREIGN KEY (asset_urn)
    REFERENCES assets(urn) ON DELETE CASCADE ON UPDATE CASCADE;

DROP INDEX IF EXISTS assets_versions_idx_urn_version;
CREATE UNIQUE INDEX IF NOT EXISTS assets_versions_idx_urn_type_service_version ON assets_versions(urn,type,service,version);

DROP INDEX IF EXISTS users_idx_uuid;
CREATE UNIQUE INDEX users_idx_uuid ON users(uuid);

DROP INDEX IF EXISTS users_idx_email;
CREATE UNIQUE INDEX users_idx_email ON users(email);

----


DROP INDEX IF EXISTS idx_assets_versions_namespace_id;
ALTER TABLE assets_versions DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_asset_owners_namespace_id;
ALTER TABLE asset_owners DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_asset_probes_namespace_id;
ALTER TABLE asset_probes DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_tags_namespace_id;
ALTER TABLE tags DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_tag_templates_namespace_id;
ALTER TABLE tag_templates DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_tag_template_fields_namespace_id;
ALTER TABLE tag_template_fields DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_users_namespace_id;
ALTER TABLE users DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_stars_namespace_id;
ALTER TABLE stars DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_lineage_graph_namespace_id;
ALTER TABLE lineage_graph DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_discussions_namespace_id;
ALTER TABLE discussions DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_comments_namespace_id;
ALTER TABLE comments DROP COLUMN IF EXISTS namespace_id;

----

DROP TABLE IF EXISTS namespaces;

COMMIT;

