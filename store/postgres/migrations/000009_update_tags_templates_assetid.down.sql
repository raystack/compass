BEGIN;

DROP INDEX IF EXISTS tags_idx_asset_id_field_id;

ALTER TABLE tags 
ADD COLUMN record_urn text NOT NULL,
ADD COLUMN record_type text NOT NULL,
DROP COLUMN asset_id;

CREATE UNIQUE INDEX tags_idx_record_urn_record_type_field_id ON tags(record_urn,record_type,field_id);

COMMIT;