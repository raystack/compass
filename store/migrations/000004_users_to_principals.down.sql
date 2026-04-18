-- Reverse: rename back and remove new columns
DROP INDEX IF EXISTS idx_principals_subject;
DROP INDEX IF EXISTS idx_principals_type;
DROP INDEX IF EXISTS idx_principals_namespace_id;
DROP INDEX IF EXISTS idx_principals_subject_namespace;

ALTER TABLE principals DROP COLUMN IF EXISTS metadata;
ALTER TABLE principals DROP COLUMN IF EXISTS subject;
ALTER TABLE principals DROP COLUMN IF EXISTS name;
ALTER TABLE principals DROP COLUMN IF EXISTS type;

DROP POLICY IF EXISTS principals_isolation_policy ON principals;
ALTER TABLE principals RENAME TO users;

CREATE POLICY users_isolation_policy ON users
    USING (namespace_id = current_setting('app.current_tenant')::UUID);

CREATE INDEX idx_users_uuid ON users (uuid) WHERE uuid IS NOT NULL AND uuid != '';
CREATE UNIQUE INDEX idx_users_email ON users (email, namespace_id) WHERE email IS NOT NULL AND email != '';
CREATE INDEX idx_users_namespace_id ON users (namespace_id);
