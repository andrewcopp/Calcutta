-- Restore the algorithm_id column to game_outcome_runs.

ALTER TABLE derived.game_outcome_runs
    ADD COLUMN IF NOT EXISTS algorithm_id uuid;

CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_algorithm_id
    ON derived.game_outcome_runs (algorithm_id);
