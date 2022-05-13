BEGIN;

DROP INDEX IF EXISTS tags_idx_record_urn_record_type_field_id;

ALTER TABLE tags 
ADD COLUMN asset_id text NOT NULL,
DROP COLUMN record_urn,
DROP COLUMN record_type;

CREATE UNIQUE INDEX tags_idx_asset_id_field_id ON tags(asset_id,field_id);

COMMIT;