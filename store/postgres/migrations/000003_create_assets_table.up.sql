CREATE TABLE assets (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    urn text NOT NULL,
    type text NOT NULL,
    service text NOT NULL,
    name text NOT NULL,
    description text,
    data jsonb,
    labels jsonb,
    version text NOT NULL,
    updated_by text NOT NULL,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW()
);

CREATE UNIQUE INDEX assets_idx_urn_type_service ON assets(urn,type,service);

CREATE TABLE asset_owners (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX asset_owners_idx_asset_id_user_id ON asset_owners(asset_id,user_id);
