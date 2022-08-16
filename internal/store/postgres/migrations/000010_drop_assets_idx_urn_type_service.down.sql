CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS assets_idx_urn_type_service ON assets(urn,type,service);
