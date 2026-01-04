DROP INDEX IF EXISTS uq_derived_predicted_market_share_run_team;
DROP INDEX IF EXISTS uq_derived_predicted_market_share_legacy_tournament_team;
DROP INDEX IF EXISTS uq_derived_predicted_market_share_legacy_calcutta_team;

DROP INDEX IF EXISTS uq_derived_predicted_game_outcomes_run_matchup;
DROP INDEX IF EXISTS uq_derived_predicted_game_outcomes_legacy_matchup;

ALTER TABLE IF EXISTS derived.predicted_market_share
    DROP COLUMN IF EXISTS run_id;

ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    DROP COLUMN IF EXISTS run_id;

 DO $$
 BEGIN
     IF NOT EXISTS (
         SELECT 1
         FROM pg_constraint
         WHERE conname = 'silver_predicted_game_outcomes_unique_matchup'
     ) THEN
         ALTER TABLE derived.predicted_game_outcomes
         ADD CONSTRAINT silver_predicted_game_outcomes_unique_matchup
         UNIQUE(tournament_id, game_id, team1_id, team2_id);
     END IF;
 END $$;

 CREATE UNIQUE INDEX IF NOT EXISTS silver_predicted_market_share_calcutta_team_uniq
     ON derived.predicted_market_share(calcutta_id, team_id)
     WHERE calcutta_id IS NOT NULL;

 CREATE UNIQUE INDEX IF NOT EXISTS silver_predicted_market_share_tournament_team_uniq
     ON derived.predicted_market_share(tournament_id, team_id)
     WHERE tournament_id IS NOT NULL;

DROP TABLE IF EXISTS derived.suite_calcutta_evaluations;
DROP TABLE IF EXISTS derived.suites;
DROP TABLE IF EXISTS derived.market_share_runs;
DROP TABLE IF EXISTS derived.game_outcome_runs;
DROP TABLE IF EXISTS derived.algorithms;
