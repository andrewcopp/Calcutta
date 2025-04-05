-- Drop tables in reverse order of creation (to handle foreign key constraints)
DROP TABLE IF EXISTS calcutta_portfolio_teams;
DROP TABLE IF EXISTS calcutta_portfolios;
DROP TABLE IF EXISTS calcutta_entry_teams;
DROP TABLE IF EXISTS calcutta_entries;
DROP TABLE IF EXISTS calcutta_rounds;
DROP TABLE IF EXISTS calcuttas;
DROP TABLE IF EXISTS tournament_teams;
DROP TABLE IF EXISTS tournaments;
DROP TABLE IF EXISTS schools;
DROP TABLE IF EXISTS users;

-- Drop the UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp"; 