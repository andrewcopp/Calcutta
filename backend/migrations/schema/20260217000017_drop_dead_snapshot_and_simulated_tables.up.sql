-- Drop lab.evaluations.simulated_calcutta_id column (FK to derived.simulated_calcuttas)
-- This column is never written by Python or Go code; always NULL.
ALTER TABLE lab.evaluations
    DROP CONSTRAINT IF EXISTS evaluations_simulated_calcutta_id_fkey;
DROP INDEX IF EXISTS idx_lab_evaluations_simulated_calcutta_id;
ALTER TABLE lab.evaluations
    DROP COLUMN IF EXISTS simulated_calcutta_id;

-- Drop 5 core.calcutta_snapshot_* tables (zero app references)
DROP TABLE IF EXISTS core.calcutta_snapshot_scoring_rules CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_payouts CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_entry_teams CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_entries CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshots CASCADE;

-- Drop 5 derived.simulated_* tables (zero app references)
-- NOT simulated_teams (actively used by run_resolver.go and simulate_tournaments)
DROP TABLE IF EXISTS derived.simulated_calcutta_scoring_rules CASCADE;
DROP TABLE IF EXISTS derived.simulated_calcutta_payouts CASCADE;
DROP TABLE IF EXISTS derived.simulated_entry_teams CASCADE;
DROP TABLE IF EXISTS derived.simulated_entries CASCADE;
DROP TABLE IF EXISTS derived.simulated_calcuttas CASCADE;
