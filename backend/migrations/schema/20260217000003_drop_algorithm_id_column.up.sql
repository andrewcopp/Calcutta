-- Drop the algorithm_id column from game_outcome_runs.
-- This column is redundant with prediction_model_id and is no longer written to.

DROP INDEX IF EXISTS derived.idx_derived_game_outcome_runs_algorithm_id;

ALTER TABLE derived.game_outcome_runs DROP COLUMN IF EXISTS algorithm_id;
