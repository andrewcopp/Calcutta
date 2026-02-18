-- Rename legacy PK/index/constraint names on surviving derived tables.
-- Uses ALTER INDEX ... RENAME TO which is metadata-only (no locks, no rewrite).

-- =============================================================================
-- PRIMARY KEYS
-- =============================================================================

-- predicted_game_outcomes: silver_ prefix
ALTER INDEX IF EXISTS silver_predicted_game_outcomes_pkey
  RENAME TO pk_derived_predicted_game_outcomes;

-- simulated_teams: silver_ prefix (also wrong table name in original)
ALTER INDEX IF EXISTS silver_simulated_tournaments_pkey
  RENAME TO pk_derived_simulated_teams;

-- simulated_tournaments: old table name prefix
ALTER INDEX IF EXISTS tournament_simulation_batches_pkey
  RENAME TO pk_derived_simulated_tournaments;

-- simulation_state_teams: old table name prefix
ALTER INDEX IF EXISTS tournament_state_snapshot_teams_pkey
  RENAME TO pk_derived_simulation_state_teams;

-- simulation_states: old table name prefix
ALTER INDEX IF EXISTS tournament_state_snapshots_pkey
  RENAME TO pk_derived_simulation_states;

-- =============================================================================
-- UNIQUE CONSTRAINTS (backed by unique indexes)
-- =============================================================================

ALTER INDEX IF EXISTS uq_analytics_tournament_state_snapshot_teams_snapshot_team
  RENAME TO uq_derived_simulation_state_teams_state_team;

ALTER INDEX IF EXISTS uq_analytics_tournament_simulation_batches_natural_key
  RENAME TO uq_derived_simulated_tournaments_natural_key;

ALTER INDEX IF EXISTS uq_analytics_simulated_tournaments_batch_sim_team
  RENAME TO uq_derived_simulated_teams_batch_sim_team;

ALTER INDEX IF EXISTS uq_analytics_simulated_tournaments_legacy_sim_team
  RENAME TO uq_derived_simulated_teams_legacy_sim_team;

-- =============================================================================
-- INDEXES: silver_ prefix
-- =============================================================================

ALTER INDEX IF EXISTS idx_silver_predicted_game_outcomes_tournament_id
  RENAME TO idx_derived_predicted_game_outcomes_tournament_id;

ALTER INDEX IF EXISTS idx_silver_simulated_tournaments_sim_id
  RENAME TO idx_derived_simulated_teams_sim_id;

ALTER INDEX IF EXISTS idx_silver_simulated_tournaments_team_id
  RENAME TO idx_derived_simulated_teams_team_id;

ALTER INDEX IF EXISTS idx_silver_simulated_tournaments_tournament_id
  RENAME TO idx_derived_simulated_teams_tournament_id;

-- =============================================================================
-- INDEXES: analytics_ prefix
-- =============================================================================

ALTER INDEX IF EXISTS idx_analytics_simulated_tournaments_batch_id
  RENAME TO idx_derived_simulated_teams_batch_id;

ALTER INDEX IF EXISTS idx_analytics_tournament_simulation_batches_snapshot_id
  RENAME TO idx_derived_simulated_tournaments_snapshot_id;

ALTER INDEX IF EXISTS idx_analytics_tournament_simulation_batches_tournament_id
  RENAME TO idx_derived_simulated_tournaments_tournament_id;

ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshot_teams_snapshot_id
  RENAME TO idx_derived_simulation_state_teams_state_id;

ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshot_teams_team_id
  RENAME TO idx_derived_simulation_state_teams_team_id;

ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshots_tournament_id
  RENAME TO idx_derived_simulation_states_tournament_id;
