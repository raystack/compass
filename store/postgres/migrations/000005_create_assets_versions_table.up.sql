CREATE TABLE assets_versions (
    id serial PRIMARY KEY,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    urn text NOT NULL,
    type text NOT NULL,
    service text NOT NULL,
    name text NOT NULL,
    description text,
    data jsonb,
    labels jsonb,
    version text NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamp,
    updated_at timestamp,
    owners jsonb,
    changelog jsonb NOT NULL
);

CREATE UNIQUE INDEX assets_versions_idx_asset_id_version ON assets_versions(asset_id,version);
CREATE UNIQUE INDEX assets_versions_idx_urn_type_service_version ON assets_versions(urn,type,service,version);