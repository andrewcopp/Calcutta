BEGIN;

-- Drop dead tables
DROP TABLE IF EXISTS derived.entry_bids CASCADE;
DROP TABLE IF EXISTS derived.detailed_investment_report CASCADE;
DROP TABLE IF EXISTS derived.optimization_runs CASCADE;

-- Drop dead views
DROP VIEW IF EXISTS derived.calcuttas;
DROP VIEW IF EXISTS derived.teams;
DROP VIEW IF EXISTS derived.tournaments;

-- Drop dead column on simulated_entries
ALTER TABLE derived.simulated_entries DROP COLUMN IF EXISTS source_candidate_id;

-- Update source_kind check constraint to remove dead 'from_candidate' value
ALTER TABLE derived.simulated_entries DROP CONSTRAINT IF EXISTS ck_derived_simulated_entries_source_kind;
ALTER TABLE derived.simulated_entries ADD CONSTRAINT ck_derived_simulated_entries_source_kind
  CHECK (source_kind = ANY (ARRAY['manual'::text, 'from_real_entry'::text]));

COMMIT;
