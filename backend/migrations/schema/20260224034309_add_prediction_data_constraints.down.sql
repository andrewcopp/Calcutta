-- Rollback: add_prediction_data_constraints
-- Created: 2026-02-24 03:43:09 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

DROP INDEX IF EXISTS compute.uq_predicted_team_values_batch_team;

ALTER TABLE IF EXISTS compute.prediction_batches
    DROP CONSTRAINT IF EXISTS chk_prediction_batches_through_round;

ALTER TABLE IF EXISTS compute.predicted_team_values
    DROP CONSTRAINT IF EXISTS chk_predicted_team_values_expected_points;

ALTER TABLE IF EXISTS compute.predicted_team_values
    DROP CONSTRAINT IF EXISTS chk_predicted_team_values_probabilities;
