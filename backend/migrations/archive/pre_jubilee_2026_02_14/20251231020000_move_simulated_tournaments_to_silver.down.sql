-- Rollback: Move simulated tournaments back from silver to bronze layer

-- Create bronze_simulated_tournaments table
CREATE TABLE IF NOT EXISTS bronze_simulated_tournaments (
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_bronze_simulated_tournaments_tournament_id 
    ON bronze_simulated_tournaments(tournament_id);
CREATE INDEX IF NOT EXISTS idx_bronze_simulated_tournaments_sim_id 
    ON bronze_simulated_tournaments(tournament_id, sim_id);
CREATE INDEX IF NOT EXISTS idx_bronze_simulated_tournaments_team_id 
    ON bronze_simulated_tournaments(team_id);

-- Migrate data back from silver to bronze
INSERT INTO bronze_simulated_tournaments 
    (tournament_id, sim_id, team_id, wins, byes, eliminated, created_at)
SELECT 
    tournament_id, sim_id, team_id, wins, byes, eliminated, created_at
FROM silver_simulated_tournaments
ON CONFLICT (tournament_id, sim_id, team_id) DO NOTHING;

-- Drop silver table
DROP TABLE IF EXISTS silver_simulated_tournaments;
