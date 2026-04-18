-- Rename users to principals and evolve schema
ALTER TABLE users RENAME TO principals;

-- Add new columns
ALTER TABLE principals ADD COLUMN type TEXT NOT NULL DEFAULT 'user';
ALTER TABLE principals ADD COLUMN name TEXT;
ALTER TABLE principals ADD COLUMN subject TEXT;
ALTER TABLE principals ADD COLUMN metadata JSONB DEFAULT '{}';

-- Migrate existing data: copy uuid to subject
UPDATE principals SET subject = uuid WHERE uuid IS NOT NULL AND uuid != '';

-- Add unique constraint on subject per namespace
CREATE UNIQUE INDEX idx_principals_subject_namespace ON principals (subject, namespace_id) WHERE subject IS NOT NULL AND subject != '';

-- Update RLS policy
DROP POLICY IF EXISTS users_isolation_policy ON principals;
DROP POLICY IF EXISTS principals_isolation_policy ON principals;
CREATE POLICY principals_isolation_policy ON principals
    USING (namespace_id = current_setting('app.current_tenant')::UUID);

-- Drop old indexes that reference 'users' naming
DROP INDEX IF EXISTS idx_users_uuid;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_namespace_id;

-- Create new indexes
CREATE INDEX idx_principals_subject ON principals (subject) WHERE subject IS NOT NULL AND subject != '';
CREATE INDEX idx_principals_type ON principals (type);
CREATE INDEX idx_principals_namespace_id ON principals (namespace_id);
