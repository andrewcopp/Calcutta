BEGIN;

-- ============================================================================
-- 6.4a: Rename legacy constraint names
-- ============================================================================

-- 1. silver_predicted_game_outcomes_pkey -> predicted_game_outcomes_pkey
--    The "silver_" prefix is a leftover from an old naming scheme.
--    Table: derived.predicted_game_outcomes
ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    RENAME CONSTRAINT silver_predicted_game_outcomes_pkey
    TO predicted_game_outcomes_pkey;

-- 2. silver_simulated_tournaments_pkey -> simulated_teams_pkey
--    This PK was on the WRONG table (derived.simulated_teams) but named
--    after simulated_tournaments. Rename to match the actual table.
ALTER TABLE IF EXISTS derived.simulated_teams
    RENAME CONSTRAINT silver_simulated_tournaments_pkey
    TO simulated_teams_pkey;

-- 3. Truncated FK names (PostgreSQL truncated these at 63 chars)
--
-- 3a. On derived.simulated_tournaments: the FK for simulation_state_id
--     Old (truncated): tournament_simulation_batches_tournament_state_snapshot_id_fkey
--     New:             simulated_tournaments_simulation_state_id_fkey
ALTER TABLE IF EXISTS derived.simulated_tournaments
    RENAME CONSTRAINT tournament_simulation_batches_tournament_state_snapshot_id_fkey
    TO simulated_tournaments_simulation_state_id_fkey;

-- 3b. On derived.simulation_state_teams: the FK for simulation_state_id
--     Old (truncated): tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey
--     New:             simulation_state_teams_simulation_state_id_fkey
ALTER TABLE IF EXISTS derived.simulation_state_teams
    RENAME CONSTRAINT tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey
    TO simulation_state_teams_simulation_state_id_fkey;

-- ============================================================================
-- 6.4b: Standardize UUID generation to uuid_generate_v4()
-- ============================================================================

-- derived.predicted_team_values uses gen_random_uuid(), switch to project standard
ALTER TABLE IF EXISTS derived.predicted_team_values
    ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- derived.prediction_batches uses gen_random_uuid(), switch to project standard
ALTER TABLE IF EXISTS derived.prediction_batches
    ALTER COLUMN id SET DEFAULT uuid_generate_v4();

COMMIT;
