BEGIN;

DROP POLICY IF EXISTS assets_isolation_policy ON assets;
DROP POLICY IF EXISTS assets_versions_isolation_policy ON assets_versions;
DROP POLICY IF EXISTS asset_owners_isolation_policy ON asset_owners;
DROP POLICY IF EXISTS asset_probes_isolation_policy ON asset_probes;
DROP POLICY IF EXISTS tags_isolation_policy ON tags;
DROP POLICY IF EXISTS tag_templates_isolation_policy ON tag_templates;
DROP POLICY IF EXISTS tag_template_fields_isolation_policy ON tag_template_fields;
DROP POLICY IF EXISTS users_isolation_policy ON users;
DROP POLICY IF EXISTS stars_isolation_policy ON stars;
DROP POLICY IF EXISTS lineage_graph_isolation_policy ON lineage_graph;
DROP POLICY IF EXISTS discussions_isolation_policy ON discussions;
DROP POLICY IF EXISTS comments_isolation_policy ON comments;

ALTER TABLE assets DISABLE ROW LEVEL SECURITY;
ALTER TABLE assets_versions DISABLE ROW LEVEL SECURITY;
ALTER TABLE asset_owners DISABLE ROW LEVEL SECURITY;
ALTER TABLE asset_probes DISABLE ROW LEVEL SECURITY;
ALTER TABLE tags DISABLE ROW LEVEL SECURITY;
ALTER TABLE tag_templates DISABLE ROW LEVEL SECURITY;
ALTER TABLE tag_template_fields DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE stars DISABLE ROW LEVEL SECURITY;
ALTER TABLE lineage_graph DISABLE ROW LEVEL SECURITY;
ALTER TABLE discussions DISABLE ROW LEVEL SECURITY;
ALTER TABLE comments DISABLE ROW LEVEL SECURITY;

COMMIT;