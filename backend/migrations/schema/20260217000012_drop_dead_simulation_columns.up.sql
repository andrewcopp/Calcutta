-- Drop dead table: entry_simulation_outcomes
DROP TABLE IF EXISTS derived.entry_simulation_outcomes CASCADE;

-- Drop 11 unused columns from derived.simulation_runs
ALTER TABLE derived.simulation_runs
    DROP COLUMN IF EXISTS our_rank,
    DROP COLUMN IF EXISTS our_mean_normalized_payout,
    DROP COLUMN IF EXISTS our_median_normalized_payout,
    DROP COLUMN IF EXISTS our_p_top1,
    DROP COLUMN IF EXISTS our_p_in_money,
    DROP COLUMN IF EXISTS total_simulations,
    DROP COLUMN IF EXISTS realized_finish_position,
    DROP COLUMN IF EXISTS realized_is_tied,
    DROP COLUMN IF EXISTS realized_in_the_money,
    DROP COLUMN IF EXISTS realized_payout_cents,
    DROP COLUMN IF EXISTS realized_total_points;

-- Drop focus_snapshot_entry_id + its index and FK from derived.simulation_runs
ALTER TABLE derived.simulation_runs
    DROP CONSTRAINT IF EXISTS fk_derived_simulation_runs_focus_snapshot_entry_id;
DROP INDEX IF EXISTS derived.idx_derived_simulation_runs_focus_snapshot_entry_id;
ALTER TABLE derived.simulation_runs
    DROP COLUMN IF EXISTS focus_snapshot_entry_id;

-- Drop highlighted_simulated_entry_id + its index and FK from derived.simulated_calcuttas
ALTER TABLE derived.simulated_calcuttas
    DROP CONSTRAINT IF EXISTS simulated_calcuttas_highlighted_simulated_entry_id_fkey;
DROP INDEX IF EXISTS derived.idx_derived_simulated_calcuttas_highlighted_simulated_entry_id;
ALTER TABLE derived.simulated_calcuttas
    DROP COLUMN IF EXISTS highlighted_simulated_entry_id;
