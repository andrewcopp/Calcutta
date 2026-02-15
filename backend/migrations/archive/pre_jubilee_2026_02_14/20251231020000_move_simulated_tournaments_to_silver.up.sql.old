-- Move simulated tournaments from bronze to silver layer
-- Simulated tournaments are derived data (Monte Carlo simulations), not raw data

-- Create new silver_simulated_tournaments table
CREATE TABLE IF NOT EXISTS silver_simulated_tournaments (
    id bigserial PRIMARY KEY,
    tournament_id bigint NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    sim_id integer NOT NULL,
    team_id bigint NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    wins integer NOT NULL,
    byes integer NOT NULL DEFAULT 0,
    eliminated boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE(tournament_id, sim_id, team_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_silver_simulated_tournaments_tournament_id 
    ON silver_simulated_tournaments(tournament_id);
CREATE INDEX IF NOT EXISTS idx_silver_simulated_tournaments_sim_id 
    ON silver_simulated_tournaments(tournament_id, sim_id);
CREATE INDEX IF NOT EXISTS idx_silver_simulated_tournaments_team_id 
    ON silver_simulated_tournaments(team_id);

-- Migrate existing data from bronze to silver
INSERT INTO silver_simulated_tournaments 
    (tournament_id, sim_id, team_id, wins, byes, eliminated, created_at)
SELECT 
    tournament_id, sim_id, team_id, wins, byes, eliminated, created_at
FROM bronze_simulated_tournaments
ON CONFLICT (tournament_id, sim_id, team_id) DO NOTHING;

-- Drop old bronze table
DROP TABLE IF EXISTS bronze_simulated_tournaments;
