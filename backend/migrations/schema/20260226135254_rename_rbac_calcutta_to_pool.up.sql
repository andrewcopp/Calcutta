-- Migration: rename_rbac_calcutta_to_pool
-- Created: 2026-02-26 13:52:54 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Rename permission keys
UPDATE core.permissions SET key = 'pool.config.write', description = 'Create and configure a pool' WHERE key = 'calcutta.config.write';
UPDATE core.permissions SET key = 'pool.read', description = 'Read pool data' WHERE key = 'calcutta.read';

-- Rename role key
UPDATE core.roles SET key = 'pool_admin', description = 'Manage a specific pool' WHERE key = 'calcutta_admin';

-- Update scope_type in grants
UPDATE core.grants SET scope_type = 'pool' WHERE scope_type = 'calcutta';
