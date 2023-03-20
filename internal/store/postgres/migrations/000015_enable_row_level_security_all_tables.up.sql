BEGIN;

ALTER TABLE assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE asset_owners ENABLE ROW LEVEL SECURITY;
ALTER TABLE asset_probes ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE tag_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE tag_template_fields ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE stars ENABLE ROW LEVEL SECURITY;
ALTER TABLE lineage_graph ENABLE ROW LEVEL SECURITY;
ALTER TABLE discussions ENABLE ROW LEVEL SECURITY;
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS assets_isolation_policy ON assets;
CREATE POLICY assets_isolation_policy on assets USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS assets_versions_isolation_policy ON assets_versions;
CREATE POLICY assets_versions_isolation_policy on assets_versions USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS asset_owners_isolation_policy ON asset_owners;
CREATE POLICY asset_owners_isolation_policy on asset_owners USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS asset_probes_isolation_policy ON asset_probes;
CREATE POLICY asset_probes_isolation_policy on asset_probes USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS tags_isolation_policy ON tags;
CREATE POLICY tags_isolation_policy on tags USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS tag_templates_isolation_policy ON tag_templates;
CREATE POLICY tag_templates_isolation_policy on tag_templates USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS tag_template_fields_isolation_policy ON tag_template_fields;
CREATE POLICY tag_template_fields_isolation_policy on tag_template_fields USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS users_isolation_policy ON users;
CREATE POLICY users_isolation_policy on users USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS stars_isolation_policy ON stars;
CREATE POLICY stars_isolation_policy on stars USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS lineage_graph_isolation_policy ON lineage_graph;
CREATE POLICY lineage_graph_isolation_policy on lineage_graph USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS discussions_isolation_policy ON discussions;
CREATE POLICY discussions_isolation_policy on discussions USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS comments_isolation_policy ON comments;
CREATE POLICY comments_isolation_policy on comments USING (namespace_id = current_setting('app.current_tenant')::UUID);

COMMIT;