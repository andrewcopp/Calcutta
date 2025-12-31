-- Rollback: Restore original analytics schema with string keys

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

-- Restore original schema (from 20251231000000_add_analytics_tables.up.sql)
-- Note: This is a simplified rollback - original data will be lost

CREATE TABLE bronze_tournaments (
    tournament_key VARCHAR(100) PRIMARY KEY,
    season INTEGER NOT NULL,
    tournament_name VARCHAR(200) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_tournaments_season ON bronze_tournaments(season);

CREATE TABLE bronze_teams (
    team_key VARCHAR(100) PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key) ON DELETE CASCADE,
    school_slug VARCHAR(100) NOT NULL,
    school_name VARCHAR(200) NOT NULL,
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER DEFAULT 0,
    kenpom_net NUMERIC(10,2),
    kenpom_o NUMERIC(10,2),
    kenpom_d NUMERIC(10,2),
    kenpom_adj_t NUMERIC(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_teams_tournament ON bronze_teams(tournament_key);
CREATE INDEX idx_bronze_teams_seed ON bronze_teams(seed);

CREATE TABLE bronze_simulated_tournaments (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key) ON DELETE CASCADE,
    sim_id INTEGER NOT NULL,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    wins INTEGER NOT NULL,
    byes INTEGER NOT NULL,
    eliminated BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_sim_tournaments_key_sim ON bronze_simulated_tournaments(tournament_key, sim_id);
CREATE INDEX idx_bronze_sim_tournaments_team ON bronze_simulated_tournaments(team_key);

CREATE TABLE bronze_calcuttas (
    calcutta_key VARCHAR(100) PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key) ON DELETE CASCADE,
    calcutta_name VARCHAR(200) NOT NULL,
    budget_points INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_calcuttas_tournament ON bronze_calcuttas(tournament_key);

CREATE TABLE bronze_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    entry_name VARCHAR(200) NOT NULL,
    bid_amount_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_entry_bids_calcutta ON bronze_entry_bids(calcutta_key);
CREATE INDEX idx_bronze_entry_bids_team ON bronze_entry_bids(team_key);

CREATE TABLE bronze_payouts (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    place INTEGER NOT NULL,
    payout_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, place)
);

CREATE INDEX idx_bronze_payouts_calcutta ON bronze_payouts(calcutta_key);

CREATE TABLE silver_predicted_game_outcomes (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key) ON DELETE CASCADE,
    game_id VARCHAR(100) NOT NULL,
    round INTEGER NOT NULL,
    team1_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    team2_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    p_team1_wins NUMERIC(10,6) NOT NULL,
    p_matchup NUMERIC(10,6) NOT NULL DEFAULT 1.0,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_key, game_id)
);

CREATE INDEX idx_silver_pred_games_tournament ON silver_predicted_game_outcomes(tournament_key);

CREATE TABLE silver_team_tournament_value (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key) ON DELETE CASCADE,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    expected_wins NUMERIC(10,4) NOT NULL,
    expected_points NUMERIC(10,4) NOT NULL,
    win_probability NUMERIC(10,6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_key, team_key)
);

CREATE INDEX idx_silver_team_value_tournament ON silver_team_tournament_value(tournament_key);
CREATE INDEX idx_silver_team_value_team ON silver_team_tournament_value(team_key);

CREATE TABLE silver_predicted_market_share (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    predicted_share NUMERIC(10,6) NOT NULL,
    predicted_price_points NUMERIC(10,2) NOT NULL,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, team_key)
);

CREATE INDEX idx_silver_pred_market_calcutta ON silver_predicted_market_share(calcutta_key);
CREATE INDEX idx_silver_pred_market_team ON silver_predicted_market_share(team_key);

CREATE TABLE silver_entry_simulation_outcomes (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    sim_id INTEGER NOT NULL,
    entry_name VARCHAR(200) NOT NULL,
    total_points_won NUMERIC(10,2) NOT NULL,
    total_payout_points NUMERIC(10,2) NOT NULL,
    roi NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_silver_entry_sims_calcutta_sim ON silver_entry_simulation_outcomes(calcutta_key, sim_id);
CREATE INDEX idx_silver_entry_sims_entry ON silver_entry_simulation_outcomes(entry_name);

CREATE TABLE silver_entry_performance (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    entry_name VARCHAR(200) NOT NULL,
    mean_payout_points NUMERIC(10,2) NOT NULL,
    median_payout_points NUMERIC(10,2) NOT NULL,
    p10_payout_points NUMERIC(10,2) NOT NULL,
    p90_payout_points NUMERIC(10,2) NOT NULL,
    win_probability NUMERIC(10,6) NOT NULL,
    mean_roi NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, entry_name)
);

CREATE INDEX idx_silver_entry_perf_calcutta ON silver_entry_performance(calcutta_key);
CREATE INDEX idx_silver_entry_perf_entry ON silver_entry_performance(entry_name);

CREATE TABLE gold_optimization_runs (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL UNIQUE,
    calcutta_key VARCHAR(100) NOT NULL REFERENCES bronze_calcuttas(calcutta_key) ON DELETE CASCADE,
    strategy VARCHAR(50) NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    budget_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_opt_runs_calcutta ON gold_optimization_runs(calcutta_key);
CREATE INDEX idx_gold_opt_runs_run_id ON gold_optimization_runs(run_id);

CREATE TABLE gold_recommended_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    recommended_bid_points INTEGER NOT NULL,
    expected_roi NUMERIC(10,4) NOT NULL,
    allocation_rank INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_rec_bids_run ON gold_recommended_entry_bids(run_id);
CREATE INDEX idx_gold_rec_bids_team ON gold_recommended_entry_bids(team_key);

CREATE TABLE gold_detailed_investment_report (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key) ON DELETE CASCADE,
    expected_points NUMERIC(10,4) NOT NULL,
    expected_market_points NUMERIC(10,2) NOT NULL,
    our_bid_points INTEGER NOT NULL,
    our_roi NUMERIC(10,4) NOT NULL,
    roi_degradation NUMERIC(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_invest_report_run ON gold_detailed_investment_report(run_id);
CREATE INDEX idx_gold_invest_report_team ON gold_detailed_investment_report(team_key);
