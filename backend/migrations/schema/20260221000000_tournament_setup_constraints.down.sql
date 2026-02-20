DROP INDEX IF EXISTS core.idx_teams_tournament_region_seed_active;

ALTER TABLE core.teams DROP CONSTRAINT IF EXISTS chk_teams_wins_range;
ALTER TABLE core.teams ADD CONSTRAINT chk_teams_wins_nonneg CHECK (wins >= 0);

ALTER TABLE core.teams DROP CONSTRAINT IF EXISTS chk_teams_byes_range;

DROP INDEX IF EXISTS core.idx_teams_tournament_school_active;
DROP INDEX IF EXISTS core.idx_tournaments_competition_season_active;
