BEGIN;

-- Restore source_kind check constraint with 'from_candidate' value
ALTER TABLE derived.simulated_entries DROP CONSTRAINT IF EXISTS ck_derived_simulated_entries_source_kind;
ALTER TABLE derived.simulated_entries ADD CONSTRAINT ck_derived_simulated_entries_source_kind
  CHECK (source_kind = ANY (ARRAY['manual'::text, 'from_real_entry'::text, 'from_candidate'::text]));

-- Restore source_candidate_id column
ALTER TABLE derived.simulated_entries ADD COLUMN IF NOT EXISTS source_candidate_id uuid;

-- Restore dead views (minimal stubs - original definitions may vary)
CREATE OR REPLACE VIEW derived.calcuttas AS
  SELECT id, name, tournament_id, budget_points, created_at, updated_at, deleted_at
  FROM core.calcuttas;

CREATE OR REPLACE VIEW derived.teams AS
  SELECT id, tournament_id, school_id, seed, region, byes, wins, eliminated, created_at, updated_at, deleted_at
  FROM core.tournament_teams;

CREATE OR REPLACE VIEW derived.tournaments AS
  SELECT id, name, season_id, starting_at, created_at, updated_at, deleted_at
  FROM core.tournaments;

-- Restore dead tables (empty stubs)
CREATE TABLE IF NOT EXISTS derived.entry_bids (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.detailed_investment_report (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.optimization_runs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

COMMIT;
