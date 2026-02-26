-- Rollback: pre_production_fixes
-- Created: 2026-02-26 10:00:00 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Reverse D7: Restore gen_random_uuid() default
ALTER TABLE compute.prediction_batches ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE compute.predicted_team_values ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Reverse D6: Drop invitation status/revoked_at consistency check
ALTER TABLE core.calcutta_invitations DROP CONSTRAINT ck_invitations_revoked_consistency;

-- Reverse D5: Drop short_name from core.schools
ALTER TABLE core.schools DROP COLUMN short_name;

-- Reverse D4: Drop entry name uniqueness index
DROP INDEX core.uq_entries_name_calcutta;

-- Reverse D2: Drop entry_fee_cents from core.calcuttas
ALTER TABLE core.calcuttas DROP CONSTRAINT ck_calcuttas_entry_fee_cents;
ALTER TABLE core.calcuttas DROP COLUMN entry_fee_cents;

-- Reverse D1: Drop status column from core.entries
ALTER TABLE core.entries DROP CONSTRAINT ck_entries_status;
ALTER TABLE core.entries DROP COLUMN status;
