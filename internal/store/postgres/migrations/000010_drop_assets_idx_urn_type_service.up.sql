-- By default, golang-migrate wraps multiple SQL statements in a transaction.
-- Dropping index concurrently is not allowed in a transaction. So the drop
-- statement needs to be the only statement in the migration.
DROP INDEX CONCURRENTLY IF EXISTS assets_idx_urn_type_service;
