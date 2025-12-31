-- Make calcutta_id optional and add tournament_id as alternative
-- This allows predictions to work during migration period when bronze_calcuttas is empty

ALTER TABLE silver_predicted_market_share
    ALTER COLUMN calcutta_id DROP NOT NULL;

ALTER TABLE silver_predicted_market_share
    ADD COLUMN tournament_id UUID REFERENCES bronze_tournaments(id) ON DELETE CASCADE;

-- Add constraint: must have either calcutta_id or tournament_id
ALTER TABLE silver_predicted_market_share
    ADD CONSTRAINT market_share_must_have_calcutta_or_tournament
    CHECK (calcutta_id IS NOT NULL OR tournament_id IS NOT NULL);

-- Update unique constraint to handle both cases
ALTER TABLE silver_predicted_market_share
    DROP CONSTRAINT IF EXISTS silver_predicted_market_share_calcutta_id_team_id_key;

-- Create partial unique indexes for both cases
CREATE UNIQUE INDEX silver_predicted_market_share_calcutta_team_uniq
    ON silver_predicted_market_share(calcutta_id, team_id)
    WHERE calcutta_id IS NOT NULL;

CREATE UNIQUE INDEX silver_predicted_market_share_tournament_team_uniq
    ON silver_predicted_market_share(tournament_id, team_id)
    WHERE tournament_id IS NOT NULL;
