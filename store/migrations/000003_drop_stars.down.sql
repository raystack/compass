-- Recreate the stars table
CREATE TABLE stars (
    id          bigserial PRIMARY KEY,
    namespace_id uuid NOT NULL REFERENCES namespaces(id),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_id   uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    created_at  timestamptz DEFAULT now(),
    updated_at  timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX idx_stars_user_entity ON stars(user_id, entity_id);
CREATE INDEX idx_stars_entity ON stars(entity_id);
CREATE INDEX idx_stars_namespace ON stars(namespace_id);

ALTER TABLE stars ENABLE ROW LEVEL SECURITY;
CREATE POLICY stars_ns ON stars USING (namespace_id = current_setting('app.current_tenant')::UUID);
