-- Rollback: remove eval-run scoped uniqueness and restore legacy uniqueness.

DROP INDEX IF EXISTS analytics.uq_analytics_entry_sim_outcomes_eval_run_entry_sim;
DROP INDEX IF EXISTS analytics.uq_analytics_entry_performance_eval_run_entry;
DROP INDEX IF EXISTS analytics.uq_analytics_entry_sim_outcomes_legacy_run_entry_sim;
DROP INDEX IF EXISTS analytics.uq_analytics_entry_performance_legacy_run_entry;

-- Restore legacy uniqueness (best-effort; may differ from original historical constraint names).
CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_sim_outcomes_run_entry_sim
ON analytics.entry_simulation_outcomes (run_id, entry_name, sim_id)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_performance_run_entry
ON analytics.entry_performance (run_id, entry_name)
WHERE deleted_at IS NULL;
