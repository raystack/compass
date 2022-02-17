CREATE TABLE assets_versions (
    id uuid NOT NULL,
    urn text NOT NULL,
    type text NOT NULL,
    service text NOT NULL,
    name text NOT NULL,
    description text,
    data jsonb,
    labels jsonb,
    version text NOT NULL,
    updated_by text NOT NULL,
    created_at timestamp,
    updated_at timestamp,
    owners jsonb,
    changelog jsonb NOT NULL,
    PRIMARY KEY(id, version)
);

CREATE UNIQUE INDEX assets_versions_idx_urn_type_service_version ON assets_versions(urn,type,service,version);