-- Revert changes to silver_predicted_market_share

DROP INDEX IF EXISTS silver_predicted_market_share_tournament_team_uniq;
DROP INDEX IF EXISTS silver_predicted_market_share_calcutta_team_uniq;

ALTER TABLE silver_predicted_market_share
    DROP CONSTRAINT IF EXISTS market_share_must_have_calcutta_or_tournament;

ALTER TABLE silver_predicted_market_share
    DROP COLUMN IF EXISTS tournament_id;

ALTER TABLE silver_predicted_market_share
    ALTER COLUMN calcutta_id SET NOT NULL;

-- Restore original unique constraint
ALTER TABLE silver_predicted_market_share
    ADD CONSTRAINT silver_predicted_market_share_calcutta_id_team_id_key
    UNIQUE (calcutta_id, team_id);
