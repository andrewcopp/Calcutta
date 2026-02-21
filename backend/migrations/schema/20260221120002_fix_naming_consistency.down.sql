BEGIN;

-- ============================================================================
-- Reverse 6.4b: Restore gen_random_uuid() defaults
-- ============================================================================

ALTER TABLE IF EXISTS derived.prediction_batches
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE IF EXISTS derived.predicted_team_values
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- ============================================================================
-- Reverse 6.4a: Restore legacy constraint names
-- ============================================================================

-- 3b. Restore truncated FK name on derived.simulation_state_teams
ALTER TABLE IF EXISTS derived.simulation_state_teams
    RENAME CONSTRAINT simulation_state_teams_simulation_state_id_fkey
    TO tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey;

-- 3a. Restore truncated FK name on derived.simulated_tournaments
ALTER TABLE IF EXISTS derived.simulated_tournaments
    RENAME CONSTRAINT simulated_tournaments_simulation_state_id_fkey
    TO tournament_simulation_batches_tournament_state_snapshot_id_fkey;

-- 2. Restore wrong-table PK name on derived.simulated_teams
ALTER TABLE IF EXISTS derived.simulated_teams
    RENAME CONSTRAINT simulated_teams_pkey
    TO silver_simulated_tournaments_pkey;

-- 1. Restore silver_ prefix PK name on derived.predicted_game_outcomes
ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    RENAME CONSTRAINT predicted_game_outcomes_pkey
    TO silver_predicted_game_outcomes_pkey;

COMMIT;
