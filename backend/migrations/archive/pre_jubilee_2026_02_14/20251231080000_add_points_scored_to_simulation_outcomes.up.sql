-- Add points_scored column to gold_entry_simulation_outcomes
ALTER TABLE gold_entry_simulation_outcomes 
ADD COLUMN IF NOT EXISTS points_scored DOUBLE PRECISION NOT NULL DEFAULT 0;
