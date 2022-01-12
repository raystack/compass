CREATE TABLE templates (
    urn text PRIMARY KEY,
    display_name text NOT NULL,
    description text NOT NULL,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE fields (
    id bigint PRIMARY KEY,
    urn text NOT NULL,
    display_name text NOT NULL,
    description text NOT NULL,
    data_type text NOT NULL,
    options text,
    required boolean NOT NULL,
    template_urn text NOT NULL REFERENCES templates(urn) ON DELETE CASCADE,
    created_at timestamp,
    updated_at timestamp
);

CREATE UNIQUE INDEX fields_idx_urn_template_urn ON fields(urn,template_urn);

CREATE TABLE tags (
    id bigint PRIMARY KEY,
    value text NOT NULL,
    record_type text NOT NULL,
    record_urn text NOT NULL,
    field_id bigint NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    created_at timestamp,
    updated_at timestamp
);

CREATE UNIQUE INDEX tags_idx_record_urn_record_type_field_id ON tags(record_urn,record_type,field_id);