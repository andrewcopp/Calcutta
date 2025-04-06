-- Remove expected_points and predicted_points from calcutta_portfolio_teams
ALTER TABLE calcutta_portfolio_teams DROP COLUMN predicted_points;
ALTER TABLE calcutta_portfolio_teams DROP COLUMN expected_points;

-- Rename actual_points back to points_earned in calcutta_portfolio_teams
ALTER TABLE calcutta_portfolio_teams RENAME COLUMN actual_points TO points_earned;

-- Remove maximum_points from calcutta_portfolios
ALTER TABLE calcutta_portfolios DROP COLUMN maximum_points;

-- Remove eliminated column from tournament_teams
ALTER TABLE tournament_teams DROP COLUMN eliminated;

-- Change calcutta_entry_teams.bid back to DECIMAL
ALTER TABLE calcutta_entry_teams ALTER COLUMN bid TYPE DECIMAL(10,2) USING bid::DECIMAL(10,2); 