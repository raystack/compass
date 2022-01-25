CREATE TABLE assets (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    urn text NOT NULL,
    type text NOT NULL,
    service text NOT NULL,
    name text NOT NULL,
    description text,
    data jsonb,
    labels jsonb,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW()
);

CREATE UNIQUE INDEX assets_idx_urn_type_service ON assets(urn,type,service);
