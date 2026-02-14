-- Rollback run-lineage spine tables for tournament simulations and calcutta evaluations.

-- Drop added columns on existing tables
DROP INDEX IF EXISTS idx_analytics_entry_performance_eval_run_id;
ALTER TABLE analytics.entry_performance
    DROP COLUMN IF EXISTS calcutta_evaluation_run_id;

DROP INDEX IF EXISTS idx_analytics_entry_simulation_outcomes_eval_run_id;
ALTER TABLE analytics.entry_simulation_outcomes
    DROP COLUMN IF EXISTS calcutta_evaluation_run_id;

DROP INDEX IF EXISTS idx_analytics_simulated_tournaments_batch_id;
ALTER TABLE analytics.simulated_tournaments
    DROP COLUMN IF EXISTS tournament_simulation_batch_id;

-- Drop new tables (children first)
DROP TABLE IF EXISTS analytics.calcutta_evaluation_runs CASCADE;
DROP TABLE IF EXISTS analytics.tournament_simulation_batches CASCADE;
DROP TABLE IF EXISTS analytics.tournament_state_snapshot_teams CASCADE;
DROP TABLE IF EXISTS analytics.tournament_state_snapshots CASCADE;
