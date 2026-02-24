-- Migration: add_prediction_data_constraints
-- Created: 2026-02-24 03:43:09 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Unique team per batch (soft-delete aware)
CREATE UNIQUE INDEX IF NOT EXISTS uq_predicted_team_values_batch_team
    ON compute.predicted_team_values (prediction_batch_id, team_id)
    WHERE deleted_at IS NULL;

-- Range: through_round 0-7
ALTER TABLE compute.prediction_batches
    ADD CONSTRAINT chk_prediction_batches_through_round
    CHECK (through_round >= 0 AND through_round <= 7);

-- Range: expected_points non-negative
ALTER TABLE compute.predicted_team_values
    ADD CONSTRAINT chk_predicted_team_values_expected_points
    CHECK (expected_points >= 0);

-- Range: round probabilities 0-1
ALTER TABLE compute.predicted_team_values
    ADD CONSTRAINT chk_predicted_team_values_probabilities
    CHECK (
        (p_round_1 IS NULL OR (p_round_1 >= 0 AND p_round_1 <= 1)) AND
        (p_round_2 IS NULL OR (p_round_2 >= 0 AND p_round_2 <= 1)) AND
        (p_round_3 IS NULL OR (p_round_3 >= 0 AND p_round_3 <= 1)) AND
        (p_round_4 IS NULL OR (p_round_4 >= 0 AND p_round_4 <= 1)) AND
        (p_round_5 IS NULL OR (p_round_5 >= 0 AND p_round_5 <= 1)) AND
        (p_round_6 IS NULL OR (p_round_6 >= 0 AND p_round_6 <= 1)) AND
        (p_round_7 IS NULL OR (p_round_7 >= 0 AND p_round_7 <= 1))
    );
