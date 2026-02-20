-- Prevent duplicate tournaments for the same competition+season (partial index for soft deletes)
CREATE UNIQUE INDEX IF NOT EXISTS idx_tournaments_competition_season_active
  ON core.tournaments (competition_id, season_id)
  WHERE deleted_at IS NULL;

-- Prevent same school appearing twice in a tournament
CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_tournament_school_active
  ON core.teams (tournament_id, school_id)
  WHERE deleted_at IS NULL;

-- Enforce valid byes range (0 or 1)
ALTER TABLE core.teams
  ADD CONSTRAINT chk_teams_byes_range CHECK (byes >= 0 AND byes <= 1);

-- Replace wins non-negative check with bounded range (0-7)
ALTER TABLE core.teams DROP CONSTRAINT IF EXISTS chk_teams_wins_nonneg;
ALTER TABLE core.teams
  ADD CONSTRAINT chk_teams_wins_range CHECK (wins >= 0 AND wins <= 7);

-- Index for bracket builder queries by tournament+region+seed
CREATE INDEX IF NOT EXISTS idx_teams_tournament_region_seed_active
  ON core.teams (tournament_id, region, seed)
  WHERE deleted_at IS NULL;
