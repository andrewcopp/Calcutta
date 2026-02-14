-- Rollback immutable calcutta snapshots.

-- Unwire calcutta evaluation runs from snapshots
ALTER TABLE analytics.calcutta_evaluation_runs
    DROP CONSTRAINT IF EXISTS fk_analytics_calcutta_evaluation_runs_calcutta_snapshot_id;

-- Drop tables (children first)
DROP TABLE IF EXISTS core.calcutta_snapshot_scoring_rules CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_payouts CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_entry_teams CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshot_entries CASCADE;
DROP TABLE IF EXISTS core.calcutta_snapshots CASCADE;
