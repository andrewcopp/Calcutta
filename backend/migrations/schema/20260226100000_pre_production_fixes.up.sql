-- Migration: pre_production_fixes
-- Created: 2026-02-26 10:00:00 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- D1: Add status column to core.entries
ALTER TABLE core.entries ADD COLUMN status text NOT NULL DEFAULT 'draft';
ALTER TABLE core.entries ADD CONSTRAINT ck_entries_status CHECK (status IN ('draft', 'submitted'));

-- Backfill: entries that already have teams are 'submitted'
UPDATE core.entries SET status = 'submitted'
WHERE id IN (SELECT DISTINCT entry_id FROM core.entry_teams WHERE deleted_at IS NULL);

-- D2: Add entry_fee_cents to core.calcuttas
ALTER TABLE core.calcuttas ADD COLUMN entry_fee_cents integer NOT NULL DEFAULT 0;
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_calcuttas_entry_fee_cents CHECK (entry_fee_cents >= 0);

-- D4: Entry name uniqueness within calcutta
CREATE UNIQUE INDEX uq_entries_name_calcutta ON core.entries (LOWER(TRIM(name)), calcutta_id) WHERE deleted_at IS NULL;

-- D5: Add short_name to core.schools
ALTER TABLE core.schools ADD COLUMN short_name varchar(50);

-- D6: Invitation status/revoked_at consistency
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT ck_invitations_revoked_consistency CHECK ((status = 'revoked') = (revoked_at IS NOT NULL));

-- D7: Standardize UUID function (tables were moved from derived to compute)
ALTER TABLE compute.predicted_team_values ALTER COLUMN id SET DEFAULT public.uuid_generate_v4();
ALTER TABLE compute.prediction_batches ALTER COLUMN id SET DEFAULT public.uuid_generate_v4();
