-- Change calcutta_entry_teams.bid from DECIMAL to INTEGER
ALTER TABLE calcutta_entry_teams ALTER COLUMN bid TYPE INTEGER USING bid::INTEGER;

-- Add eliminated column to tournament_teams
ALTER TABLE tournament_teams ADD COLUMN eliminated BOOLEAN NOT NULL DEFAULT FALSE;

-- Add maximum_points to calcutta_portfolios
ALTER TABLE calcutta_portfolios ADD COLUMN maximum_points DECIMAL(10,2) NOT NULL DEFAULT 0;

-- Rename points_earned to actual_points in calcutta_portfolio_teams
ALTER TABLE calcutta_portfolio_teams RENAME COLUMN points_earned TO actual_points;

-- Add expected_points and predicted_points to calcutta_portfolio_teams
ALTER TABLE calcutta_portfolio_teams ADD COLUMN expected_points DECIMAL(10,2) NOT NULL DEFAULT 0;
ALTER TABLE calcutta_portfolio_teams ADD COLUMN predicted_points DECIMAL(10,2) NOT NULL DEFAULT 0; 