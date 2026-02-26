-- Rollback: rename_rbac_calcutta_to_pool
-- Created: 2026-02-26 13:52:54 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Revert scope_type in grants
UPDATE core.grants SET scope_type = 'calcutta' WHERE scope_type = 'pool';

-- Revert role key
UPDATE core.roles SET key = 'calcutta_admin', description = 'Manage a specific calcutta' WHERE key = 'pool_admin';

-- Revert permission keys
UPDATE core.permissions SET key = 'calcutta.read', description = 'Read calcutta data' WHERE key = 'pool.read';
UPDATE core.permissions SET key = 'calcutta.config.write', description = 'Create and configure a calcutta' WHERE key = 'pool.config.write';
