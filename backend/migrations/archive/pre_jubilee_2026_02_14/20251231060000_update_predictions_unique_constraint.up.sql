-- Update UNIQUE constraint on silver_predicted_game_outcomes
-- to allow multiple matchup predictions per game_id

-- Drop old constraint
ALTER TABLE silver_predicted_game_outcomes 
DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_tournament_id_game_id_key;

-- Add new constraint that includes team matchups
ALTER TABLE silver_predicted_game_outcomes 
ADD CONSTRAINT silver_predicted_game_outcomes_unique_matchup 
UNIQUE(tournament_id, game_id, team1_id, team2_id);
