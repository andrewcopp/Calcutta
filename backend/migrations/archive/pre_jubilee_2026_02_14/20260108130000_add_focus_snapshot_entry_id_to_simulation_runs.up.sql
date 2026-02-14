ALTER TABLE IF EXISTS derived.simulation_runs
    ADD COLUMN IF NOT EXISTS focus_snapshot_entry_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_derived_simulation_runs_focus_snapshot_entry_id'
    ) THEN
        ALTER TABLE derived.simulation_runs
            ADD CONSTRAINT fk_derived_simulation_runs_focus_snapshot_entry_id
            FOREIGN KEY (focus_snapshot_entry_id)
            REFERENCES core.calcutta_snapshot_entries(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_focus_snapshot_entry_id
ON derived.simulation_runs(focus_snapshot_entry_id)
WHERE deleted_at IS NULL;
