-- =============================================================================
-- Migration: 20260223000002_rename_simulation_states_to_tournament_snapshots
-- Renames derived.simulation_states -> derived.tournament_snapshots
-- Renames derived.simulation_state_teams -> derived.tournament_snapshot_teams
-- =============================================================================

-- 1. Rename tables
ALTER TABLE derived.simulation_states RENAME TO tournament_snapshots;
ALTER TABLE derived.simulation_state_teams RENAME TO tournament_snapshot_teams;

-- 2. Rename columns
ALTER TABLE derived.tournament_snapshot_teams RENAME COLUMN simulation_state_id TO tournament_snapshot_id;
ALTER TABLE derived.simulated_tournaments RENAME COLUMN simulation_state_id TO tournament_snapshot_id;

-- 3. Rename primary keys
ALTER TABLE derived.tournament_snapshots RENAME CONSTRAINT simulation_states_pkey TO tournament_snapshots_pkey;
ALTER TABLE derived.tournament_snapshot_teams RENAME CONSTRAINT simulation_state_teams_pkey TO tournament_snapshot_teams_pkey;

-- 4. Rename unique constraint
ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT uq_derived_simulation_state_teams_state_team
  TO uq_derived_tournament_snapshot_teams_snapshot_team;

-- 5. Rename foreign keys
ALTER TABLE derived.tournament_snapshots
  RENAME CONSTRAINT simulation_states_tournament_id_fkey
  TO tournament_snapshots_tournament_id_fkey;

ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT simulation_state_teams_simulation_state_id_fkey
  TO tournament_snapshot_teams_tournament_snapshot_id_fkey;

ALTER TABLE derived.tournament_snapshot_teams
  RENAME CONSTRAINT simulation_state_teams_team_id_fkey
  TO tournament_snapshot_teams_team_id_fkey;

ALTER TABLE derived.simulated_tournaments
  RENAME CONSTRAINT simulated_tournaments_simulation_state_id_fkey
  TO simulated_tournaments_tournament_snapshot_id_fkey;

-- 6. Rename indexes
ALTER INDEX idx_analytics_tournament_state_snapshots_tournament_id
  RENAME TO idx_derived_tournament_snapshots_tournament_id;

ALTER INDEX idx_analytics_tournament_state_snapshot_teams_snapshot_id
  RENAME TO idx_derived_tournament_snapshot_teams_snapshot_id;

ALTER INDEX idx_analytics_tournament_state_snapshot_teams_team_id
  RENAME TO idx_derived_tournament_snapshot_teams_team_id;

ALTER INDEX idx_analytics_tournament_simulation_batches_snapshot_id
  RENAME TO idx_derived_simulated_tournaments_snapshot_id;

-- 7. Rename triggers (DROP + CREATE)
DROP TRIGGER IF EXISTS trg_derived_simulation_states_updated_at ON derived.tournament_snapshots;
CREATE TRIGGER trg_derived_tournament_snapshots_updated_at
  BEFORE UPDATE ON derived.tournament_snapshots
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_derived_simulation_state_teams_updated_at ON derived.tournament_snapshot_teams;
CREATE TRIGGER trg_derived_tournament_snapshot_teams_updated_at
  BEFORE UPDATE ON derived.tournament_snapshot_teams
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
