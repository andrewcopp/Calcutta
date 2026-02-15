-- Drop obsolete unique indexes that prevent multiple market-share runs.
-- The prediction registry migration introduces run-scoped uniqueness on (run_id, team_id).
-- These legacy indexes enforce uniqueness on (tournament_id, team_id) and (calcutta_id, team_id)
-- regardless of run_id and therefore block multiple runs.

DROP INDEX IF EXISTS derived.silver_predicted_market_share_calcutta_team_uniq;
DROP INDEX IF EXISTS derived.silver_predicted_market_share_tournament_team_uniq;
