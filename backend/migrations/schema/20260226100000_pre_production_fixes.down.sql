-- Rollback: pre_production_fixes
-- Created: 2026-02-26 10:00:00 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Reverse D7: Restore gen_random_uuid() default
ALTER TABLE compute.prediction_batches ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE compute.predicted_team_values ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Reverse D6: Drop invitation status/revoked_at consistency check
ALTER TABLE core.pool_invitations DROP CONSTRAINT ck_invitations_revoked_consistency;

-- Reverse D5: Drop short_name from core.schools
ALTER TABLE core.schools DROP COLUMN short_name;

-- Reverse D4: Drop portfolio name uniqueness index
DROP INDEX core.uq_portfolios_name_pool;

-- Reverse D2: Drop entry_fee_cents from core.pools
ALTER TABLE core.pools DROP CONSTRAINT ck_pools_entry_fee_cents;
ALTER TABLE core.pools DROP COLUMN entry_fee_cents;

-- Reverse D1: Drop status column from core.portfolios
ALTER TABLE core.portfolios DROP CONSTRAINT ck_portfolios_status;
ALTER TABLE core.portfolios DROP COLUMN status;
