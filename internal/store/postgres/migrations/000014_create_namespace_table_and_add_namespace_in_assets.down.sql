BEGIN;

DROP INDEX CONCURRENTLY IF EXISTS idx_assets_namespace_id;
ALTER TABLE assets DROP COLUMN namespace_id;
DROP TABLE IF EXISTS namespaces;

COMMIT;

