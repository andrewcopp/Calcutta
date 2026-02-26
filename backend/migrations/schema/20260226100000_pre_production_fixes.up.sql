-- Migration: pre_production_fixes
-- Created: 2026-02-26 10:00:00 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- D1: Add status column to core.portfolios (was core.entries)
ALTER TABLE core.portfolios ADD COLUMN status text NOT NULL DEFAULT 'draft';
ALTER TABLE core.portfolios ADD CONSTRAINT ck_portfolios_status CHECK (status IN ('draft', 'submitted'));

-- Backfill: portfolios that already have investments are 'submitted'
UPDATE core.portfolios SET status = 'submitted'
WHERE id IN (SELECT DISTINCT portfolio_id FROM core.investments WHERE deleted_at IS NULL);

-- D2: Add entry_fee_cents to core.pools (was core.calcuttas)
ALTER TABLE core.pools ADD COLUMN entry_fee_cents integer NOT NULL DEFAULT 0;
ALTER TABLE core.pools ADD CONSTRAINT ck_pools_entry_fee_cents CHECK (entry_fee_cents >= 0);

-- D4: Portfolio name uniqueness within pool (was entry name uniqueness within calcutta)
CREATE UNIQUE INDEX uq_portfolios_name_pool ON core.portfolios (LOWER(TRIM(name)), pool_id) WHERE deleted_at IS NULL;

-- D5: Add short_name to core.schools
ALTER TABLE core.schools ADD COLUMN short_name varchar(50);

-- D6: Invitation status/revoked_at consistency
ALTER TABLE core.pool_invitations ADD CONSTRAINT ck_invitations_revoked_consistency CHECK ((status = 'revoked') = (revoked_at IS NOT NULL));

-- D7: Standardize UUID function (tables were moved from derived to compute)
ALTER TABLE compute.predicted_team_values ALTER COLUMN id SET DEFAULT public.uuid_generate_v4();
ALTER TABLE compute.prediction_batches ALTER COLUMN id SET DEFAULT public.uuid_generate_v4();
