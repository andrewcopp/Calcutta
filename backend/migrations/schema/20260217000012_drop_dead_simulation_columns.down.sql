-- Restore columns on derived.simulation_runs
ALTER TABLE derived.simulation_runs
    ADD COLUMN IF NOT EXISTS our_rank integer,
    ADD COLUMN IF NOT EXISTS our_mean_normalized_payout double precision,
    ADD COLUMN IF NOT EXISTS our_median_normalized_payout double precision,
    ADD COLUMN IF NOT EXISTS our_p_top1 double precision,
    ADD COLUMN IF NOT EXISTS our_p_in_money double precision,
    ADD COLUMN IF NOT EXISTS total_simulations integer,
    ADD COLUMN IF NOT EXISTS realized_finish_position integer,
    ADD COLUMN IF NOT EXISTS realized_is_tied boolean,
    ADD COLUMN IF NOT EXISTS realized_in_the_money boolean,
    ADD COLUMN IF NOT EXISTS realized_payout_cents integer,
    ADD COLUMN IF NOT EXISTS realized_total_points double precision;

-- Restore focus_snapshot_entry_id
ALTER TABLE derived.simulation_runs
    ADD COLUMN IF NOT EXISTS focus_snapshot_entry_id uuid;
ALTER TABLE derived.simulation_runs
    ADD CONSTRAINT fk_derived_simulation_runs_focus_snapshot_entry_id
    FOREIGN KEY (focus_snapshot_entry_id) REFERENCES core.calcutta_snapshot_entries(id);
CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_focus_snapshot_entry_id
    ON derived.simulation_runs (focus_snapshot_entry_id) WHERE (deleted_at IS NULL);

-- Restore highlighted_simulated_entry_id
ALTER TABLE derived.simulated_calcuttas
    ADD COLUMN IF NOT EXISTS highlighted_simulated_entry_id uuid;
ALTER TABLE derived.simulated_calcuttas
    ADD CONSTRAINT simulated_calcuttas_highlighted_simulated_entry_id_fkey
    FOREIGN KEY (highlighted_simulated_entry_id) REFERENCES derived.simulated_entries(id);
CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcuttas_highlighted_simulated_entry_id
    ON derived.simulated_calcuttas (highlighted_simulated_entry_id) WHERE (deleted_at IS NULL);

-- Note: entry_simulation_outcomes table is not restored here as it was dropped with CASCADE.
-- If needed, recreate from the jubilee baseline migration.
