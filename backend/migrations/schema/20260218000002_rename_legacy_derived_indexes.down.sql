-- Revert renamed indexes back to legacy names.

-- PRIMARY KEYS
ALTER INDEX IF EXISTS pk_derived_predicted_game_outcomes
  RENAME TO silver_predicted_game_outcomes_pkey;

ALTER INDEX IF EXISTS pk_derived_simulated_teams
  RENAME TO silver_simulated_tournaments_pkey;

ALTER INDEX IF EXISTS pk_derived_simulated_tournaments
  RENAME TO tournament_simulation_batches_pkey;

ALTER INDEX IF EXISTS pk_derived_simulation_state_teams
  RENAME TO tournament_state_snapshot_teams_pkey;

ALTER INDEX IF EXISTS pk_derived_simulation_states
  RENAME TO tournament_state_snapshots_pkey;

-- UNIQUE CONSTRAINTS
ALTER INDEX IF EXISTS uq_derived_simulation_state_teams_state_team
  RENAME TO uq_analytics_tournament_state_snapshot_teams_snapshot_team;

ALTER INDEX IF EXISTS uq_derived_simulated_tournaments_natural_key
  RENAME TO uq_analytics_tournament_simulation_batches_natural_key;

ALTER INDEX IF EXISTS uq_derived_simulated_teams_batch_sim_team
  RENAME TO uq_analytics_simulated_tournaments_batch_sim_team;

ALTER INDEX IF EXISTS uq_derived_simulated_teams_legacy_sim_team
  RENAME TO uq_analytics_simulated_tournaments_legacy_sim_team;

-- INDEXES: formerly silver_
ALTER INDEX IF EXISTS idx_derived_predicted_game_outcomes_tournament_id
  RENAME TO idx_silver_predicted_game_outcomes_tournament_id;

ALTER INDEX IF EXISTS idx_derived_simulated_teams_sim_id
  RENAME TO idx_silver_simulated_tournaments_sim_id;

ALTER INDEX IF EXISTS idx_derived_simulated_teams_team_id
  RENAME TO idx_silver_simulated_tournaments_team_id;

ALTER INDEX IF EXISTS idx_derived_simulated_teams_tournament_id
  RENAME TO idx_silver_simulated_tournaments_tournament_id;

-- INDEXES: formerly analytics_
ALTER INDEX IF EXISTS idx_derived_simulated_teams_batch_id
  RENAME TO idx_analytics_simulated_tournaments_batch_id;

ALTER INDEX IF EXISTS idx_derived_simulated_tournaments_snapshot_id
  RENAME TO idx_analytics_tournament_simulation_batches_snapshot_id;

ALTER INDEX IF EXISTS idx_derived_simulated_tournaments_tournament_id
  RENAME TO idx_analytics_tournament_simulation_batches_tournament_id;

ALTER INDEX IF EXISTS idx_derived_simulation_state_teams_state_id
  RENAME TO idx_analytics_tournament_state_snapshot_teams_snapshot_id;

ALTER INDEX IF EXISTS idx_derived_simulation_state_teams_team_id
  RENAME TO idx_analytics_tournament_state_snapshot_teams_team_id;

ALTER INDEX IF EXISTS idx_derived_simulation_states_tournament_id
  RENAME TO idx_analytics_tournament_state_snapshots_tournament_id;
