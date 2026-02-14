-- Rollback: Drop new evaluation tables and views

-- Drop backward compatibility views
DROP VIEW IF EXISTS derived.v_algorithms;
DROP VIEW IF EXISTS derived.v_strategy_generation_run_bids;

-- Remove prediction_model_id from game_outcome_runs
ALTER TABLE derived.game_outcome_runs
    DROP COLUMN IF EXISTS prediction_model_id;

-- Drop optimized_entries table
DROP TABLE IF EXISTS derived.optimized_entries;

-- Drop prediction_models table
DROP TABLE IF EXISTS derived.prediction_models;
