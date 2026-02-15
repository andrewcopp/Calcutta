DROP INDEX IF EXISTS idx_derived_simulation_runs_focus_snapshot_entry_id;

ALTER TABLE IF EXISTS derived.simulation_runs
    DROP CONSTRAINT IF EXISTS fk_derived_simulation_runs_focus_snapshot_entry_id;

ALTER TABLE IF EXISTS derived.simulation_runs
    DROP COLUMN IF EXISTS focus_snapshot_entry_id;
