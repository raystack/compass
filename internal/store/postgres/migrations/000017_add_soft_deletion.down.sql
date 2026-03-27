ALTER TABLE assets DROP COLUMN IF EXISTS refreshed_at;
ALTER TABLE assets DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE assets_versions DROP COLUMN IF EXISTS is_deleted;
