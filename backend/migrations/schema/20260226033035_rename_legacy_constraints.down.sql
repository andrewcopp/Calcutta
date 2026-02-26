-- Rollback: rename_legacy_constraints
-- Created: 2026-02-26 03:30:35 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Reverse primary key constraint renames
ALTER TABLE compute.predicted_game_outcomes RENAME CONSTRAINT predicted_game_outcomes_pkey TO silver_predicted_game_outcomes_pkey;
ALTER TABLE compute.simulated_teams RENAME CONSTRAINT simulated_teams_pkey TO silver_simulated_tournaments_pkey;
ALTER TABLE compute.simulated_tournaments RENAME CONSTRAINT simulated_tournaments_pkey TO tournament_simulation_batches_pkey;
ALTER TABLE compute.tournament_snapshot_teams RENAME CONSTRAINT tournament_snapshot_teams_pkey TO tournament_state_snapshot_teams_pkey;
ALTER TABLE compute.tournament_snapshots RENAME CONSTRAINT tournament_snapshots_pkey TO tournament_state_snapshots_pkey;

-- Reverse index renames
ALTER INDEX compute.idx_simulated_teams_simulated_tournament_id RENAME TO idx_analytics_simulated_tournaments_batch_id;
ALTER INDEX compute.idx_simulated_tournaments_tournament_id RENAME TO idx_analytics_tournament_simulation_batches_tournament_id;
ALTER INDEX compute.idx_predicted_game_outcomes_tournament_id RENAME TO idx_silver_predicted_game_outcomes_tournament_id;
ALTER INDEX compute.idx_simulated_teams_tournament_sim_id RENAME TO idx_silver_simulated_tournaments_sim_id;
ALTER INDEX compute.idx_simulated_teams_team_id RENAME TO idx_silver_simulated_tournaments_team_id;
ALTER INDEX compute.idx_simulated_teams_tournament_id RENAME TO idx_silver_simulated_tournaments_tournament_id;
