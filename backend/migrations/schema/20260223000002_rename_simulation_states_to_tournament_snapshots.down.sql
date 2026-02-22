-- =============================================================================
-- Rollback: 20260223000002_rename_simulation_states_to_tournament_snapshots
-- =============================================================================

-- 7. Restore triggers
DROP TRIGGER IF EXISTS trg_derived_tournament_snapshot_teams_updated_at ON derived.tournament_snapshot_teams;
CREATE TRIGGER trg_derived_simulation_state_teams_updated_at
  BEFORE UPDATE ON derived.tournament_snapshot_teams
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_derived_tournament_snapshots_updated_at ON derived.tournament_snapshots;
CREATE TRIGGER trg_derived_simulation_states_updated_at
  BEFORE UPDATE ON derived.tournament_snapshots
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 6. Restore indexes
ALTER INDEX idx_derived_simulated_tournaments_snapshot_id
  RENAME TO idx_analytics_tournament_simulation_batches_snapshot_id;

ALTER INDEX idx_derived_tournament_snapshot_teams_team_id
  RENAME TO idx_analytics_tournament_state_snapshot_teams_team_id;

ALTER INDEX idx_derived_tournament_snapshot_teams_snapshot_id
  RENAME TO idx_analytics_tournament_state_snapshot_teams_snapshot_id;

ALTER INDEX idx_derived_tournament_snapshots_tournament_id
  RENAME TO idx_analytics_tournament_state_snapshots_tournament_id;

-- 5. Restore foreign keys
ALTER TABLE derived.simulated_tournaments
  RENAME CONSTRAINT simulated_tournaments_tournament_snapshot_id_fkey
  TO simulated_tournaments_simulation_state_id_fkey;

ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT tournament_snapshot_teams_team_id_fkey
  TO simulation_state_teams_team_id_fkey;

ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT tournament_snapshot_teams_tournament_snapshot_id_fkey
  TO simulation_state_teams_simulation_state_id_fkey;

ALTER TABLE derived.tournament_snapshots
  RENAME CONSTRAINT tournament_snapshots_tournament_id_fkey
  TO simulation_states_tournament_id_fkey;

-- 4. Restore unique constraint
ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT uq_derived_tournament_snapshot_teams_snapshot_team
  TO uq_derived_simulation_state_teams_state_team;

-- 3. Restore primary keys
ALTER TABLE derived.tournament_snapshot_teams RENAME CONSTRAINT tournament_snapshot_teams_pkey TO simulation_state_teams_pkey;
ALTER TABLE derived.tournament_snapshots RENAME CONSTRAINT tournament_snapshots_pkey TO simulation_states_pkey;

-- 2. Restore columns
ALTER TABLE derived.simulated_tournaments RENAME COLUMN tournament_snapshot_id TO simulation_state_id;
ALTER TABLE derived.tournament_snapshot_teams RENAME COLUMN tournament_snapshot_id TO simulation_state_id;

-- 1. Restore tables
ALTER TABLE derived.tournament_snapshot_teams RENAME TO simulation_state_teams;
ALTER TABLE derived.tournament_snapshots RENAME TO simulation_states;
