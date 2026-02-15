-- Revert UNIQUE constraint change on silver_predicted_game_outcomes

-- Drop new constraint
ALTER TABLE silver_predicted_game_outcomes 
DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_unique_matchup;

-- Restore old constraint
ALTER TABLE silver_predicted_game_outcomes 
ADD CONSTRAINT silver_predicted_game_outcomes_tournament_id_game_id_key 
UNIQUE(tournament_id, game_id);
