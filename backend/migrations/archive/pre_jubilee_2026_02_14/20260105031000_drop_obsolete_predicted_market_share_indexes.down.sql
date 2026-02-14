-- Recreate obsolete unique indexes (legacy behavior).

CREATE UNIQUE INDEX IF NOT EXISTS silver_predicted_market_share_calcutta_team_uniq
    ON derived.predicted_market_share(calcutta_id, team_id)
    WHERE calcutta_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS silver_predicted_market_share_tournament_team_uniq
    ON derived.predicted_market_share(tournament_id, team_id)
    WHERE tournament_id IS NOT NULL;
