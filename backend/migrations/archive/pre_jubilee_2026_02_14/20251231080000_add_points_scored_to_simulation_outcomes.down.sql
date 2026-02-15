-- Remove points_scored column from gold_entry_simulation_outcomes
ALTER TABLE gold_entry_simulation_outcomes 
DROP COLUMN IF EXISTS points_scored;
