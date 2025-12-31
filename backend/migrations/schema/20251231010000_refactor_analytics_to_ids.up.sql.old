-- Migration: Refactor analytics tables to use integer IDs instead of string keys
-- This is a complete rebuild of the analytics schema for cleaner, more efficient design

-- Drop existing analytics tables (they only have test data)
DROP TABLE IF EXISTS gold_detailed_investment_report CASCADE;
DROP TABLE IF EXISTS gold_recommended_entry_bids CASCADE;
DROP TABLE IF EXISTS gold_optimization_runs CASCADE;
DROP TABLE IF EXISTS silver_entry_performance CASCADE;
DROP TABLE IF EXISTS silver_entry_simulation_outcomes CASCADE;
DROP TABLE IF EXISTS silver_predicted_market_share CASCADE;
DROP TABLE IF EXISTS silver_team_tournament_value CASCADE;
DROP TABLE IF EXISTS silver_predicted_game_outcomes CASCADE;
DROP TABLE IF EXISTS bronze_payouts CASCADE;
DROP TABLE IF EXISTS bronze_entry_bids CASCADE;
DROP TABLE IF EXISTS bronze_calcuttas CASCADE;
DROP TABLE IF EXISTS bronze_simulated_tournaments CASCADE;
DROP TABLE IF EXISTS bronze_teams CASCADE;
DROP TABLE IF EXISTS bronze_tournaments CASCADE;

-- ============================================================================
-- BRONZE LAYER: Raw simulation and tournament data
-- ============================================================================

-- Tournaments (one per season/year)
CREATE TABLE bronze_tournaments (
    id BIGSERIAL PRIMARY KEY,
    season INTEGER NOT NULL,
    tournament_name VARCHAR(200) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(season, tournament_name)
);

CREATE INDEX idx_bronze_tournaments_season ON bronze_tournaments(season);

-- Teams in tournaments
CREATE TABLE bronze_teams (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    school_slug VARCHAR(100) NOT NULL,
    school_name VARCHAR(200) NOT NULL,
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER DEFAULT 0,
    kenpom_net NUMERIC(10,2),
    kenpom_o NUMERIC(10,2),
    kenpom_d NUMERIC(10,2),
    kenpom_adj_t NUMERIC(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, school_slug)
);

CREATE INDEX idx_bronze_teams_tournament ON bronze_teams(tournament_id);
CREATE INDEX idx_bronze_teams_seed ON bronze_teams(seed);
CREATE INDEX idx_bronze_teams_school_slug ON bronze_teams(school_slug);

-- Simulated tournament outcomes (Monte Carlo results)
CREATE TABLE bronze_simulated_tournaments (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    sim_id INTEGER NOT NULL,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    wins INTEGER NOT NULL,
    byes INTEGER NOT NULL,
    eliminated BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_sim_tournaments_tournament_sim ON bronze_simulated_tournaments(tournament_id, sim_id);
CREATE INDEX idx_bronze_sim_tournaments_team ON bronze_simulated_tournaments(team_id);

-- Calcutta auction metadata
CREATE TABLE bronze_calcuttas (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    calcutta_name VARCHAR(200) NOT NULL,
    budget_points INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, calcutta_name)
);

CREATE INDEX idx_bronze_calcuttas_tournament ON bronze_calcuttas(tournament_id);

-- Entry bids in calcutta auctions
CREATE TABLE bronze_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    entry_name VARCHAR(200) NOT NULL,
    bid_amount_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_entry_bids_calcutta ON bronze_entry_bids(calcutta_id);
CREATE INDEX idx_bronze_entry_bids_team ON bronze_entry_bids(team_id);
CREATE INDEX idx_bronze_entry_bids_entry_name ON bronze_entry_bids(entry_name);

-- Payout structure for calcutta
CREATE TABLE bronze_payouts (
    id BIGSERIAL PRIMARY KEY,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    place INTEGER NOT NULL,
    payout_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_id, place)
);

CREATE INDEX idx_bronze_payouts_calcutta ON bronze_payouts(calcutta_id);

-- ============================================================================
-- SILVER LAYER: ML predictions and enriched data
-- ============================================================================

-- Predicted game outcomes (ML model predictions)
CREATE TABLE silver_predicted_game_outcomes (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    game_id VARCHAR(100) NOT NULL,
    round INTEGER NOT NULL, -- Inverted: 0=championship, 1=final_four, etc.
    team1_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    team2_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    p_team1_wins NUMERIC(10,6) NOT NULL,
    p_matchup NUMERIC(10,6) NOT NULL DEFAULT 1.0,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, game_id)
);

CREATE INDEX idx_silver_pred_games_tournament ON silver_predicted_game_outcomes(tournament_id);
CREATE INDEX idx_silver_pred_games_round ON silver_predicted_game_outcomes(round);

-- Team tournament value (expected points from simulations)
CREATE TABLE silver_team_tournament_value (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    expected_wins NUMERIC(10,4) NOT NULL,
    expected_points NUMERIC(10,4) NOT NULL,
    win_probability NUMERIC(10,6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, team_id)
);

CREATE INDEX idx_silver_team_value_tournament ON silver_team_tournament_value(tournament_id);
CREATE INDEX idx_silver_team_value_team ON silver_team_tournament_value(team_id);

-- Predicted market share (auction price predictions)
CREATE TABLE silver_predicted_market_share (
    id BIGSERIAL PRIMARY KEY,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    predicted_share NUMERIC(10,6) NOT NULL,
    predicted_price_points NUMERIC(10,2) NOT NULL,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_id, team_id)
);

CREATE INDEX idx_silver_pred_market_calcutta ON silver_predicted_market_share(calcutta_id);
CREATE INDEX idx_silver_pred_market_team ON silver_predicted_market_share(team_id);

-- Entry simulation outcomes (per-entry Monte Carlo results)
CREATE TABLE silver_entry_simulation_outcomes (
    id BIGSERIAL PRIMARY KEY,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    sim_id INTEGER NOT NULL,
    entry_name VARCHAR(200) NOT NULL,
    total_points_won NUMERIC(10,2) NOT NULL,
    total_payout_points NUMERIC(10,2) NOT NULL,
    roi NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_silver_entry_sims_calcutta_sim ON silver_entry_simulation_outcomes(calcutta_id, sim_id);
CREATE INDEX idx_silver_entry_sims_entry ON silver_entry_simulation_outcomes(entry_name);

-- Entry performance summary (aggregated across simulations)
CREATE TABLE silver_entry_performance (
    id BIGSERIAL PRIMARY KEY,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    entry_name VARCHAR(200) NOT NULL,
    mean_payout_points NUMERIC(10,2) NOT NULL,
    median_payout_points NUMERIC(10,2) NOT NULL,
    p10_payout_points NUMERIC(10,2) NOT NULL,
    p90_payout_points NUMERIC(10,2) NOT NULL,
    win_probability NUMERIC(10,6) NOT NULL,
    mean_roi NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_id, entry_name)
);

CREATE INDEX idx_silver_entry_perf_calcutta ON silver_entry_performance(calcutta_id);
CREATE INDEX idx_silver_entry_perf_entry ON silver_entry_performance(entry_name);

-- ============================================================================
-- GOLD LAYER: Optimization results and recommendations
-- ============================================================================

-- Optimization run metadata
CREATE TABLE gold_optimization_runs (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL UNIQUE,
    calcutta_id BIGINT NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    strategy VARCHAR(50) NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    budget_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_opt_runs_calcutta ON gold_optimization_runs(calcutta_id);
CREATE INDEX idx_gold_opt_runs_run_id ON gold_optimization_runs(run_id);

-- Recommended entry bids (optimization output)
CREATE TABLE gold_recommended_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    recommended_bid_points INTEGER NOT NULL,
    expected_roi NUMERIC(10,4) NOT NULL,
    allocation_rank INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_rec_bids_run ON gold_recommended_entry_bids(run_id);
CREATE INDEX idx_gold_rec_bids_team ON gold_recommended_entry_bids(team_id);

-- Detailed investment report (team-level analysis)
CREATE TABLE gold_detailed_investment_report (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    expected_points NUMERIC(10,4) NOT NULL,
    expected_market_points NUMERIC(10,2) NOT NULL,
    our_bid_points INTEGER NOT NULL,
    our_roi NUMERIC(10,4) NOT NULL,
    roi_degradation NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_invest_report_run ON gold_detailed_investment_report(run_id);
CREATE INDEX idx_gold_invest_report_team ON gold_detailed_investment_report(team_id);
