-- Calcutta Analytics Database Schema
-- Medallion Architecture: Bronze (raw) -> Silver (cleaned) -> Gold (business metrics)
-- Designed for Airflow orchestration with polyglot microservices

-- ============================================================================
-- BRONZE LAYER: Raw simulation data
-- ============================================================================

-- Tournament metadata
CREATE TABLE bronze_tournaments (
    tournament_key VARCHAR(100) PRIMARY KEY,
    season INTEGER NOT NULL,
    tournament_name VARCHAR(200) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_tournaments_season ON bronze_tournaments(season);

-- Team information
CREATE TABLE bronze_teams (
    team_key VARCHAR(100) PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key),
    school_slug VARCHAR(100) NOT NULL,
    school_name VARCHAR(200) NOT NULL,
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER DEFAULT 0,
    kenpom_net DECIMAL(10,2),
    kenpom_o DECIMAL(10,2),
    kenpom_d DECIMAL(10,2),
    kenpom_adj_t DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_teams_tournament ON bronze_teams(tournament_key);
CREATE INDEX idx_bronze_teams_seed ON bronze_teams(seed);

-- Simulated tournament outcomes (one row per team per simulation)
CREATE TABLE bronze_simulated_tournaments (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key),
    sim_id INTEGER NOT NULL,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    wins INTEGER NOT NULL,
    byes INTEGER NOT NULL,
    eliminated BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_sim_tournaments_key_sim ON bronze_simulated_tournaments(tournament_key, sim_id);
CREATE INDEX idx_bronze_sim_tournaments_team ON bronze_simulated_tournaments(team_key);

-- Calcutta auction metadata
CREATE TABLE bronze_calcuttas (
    calcutta_key VARCHAR(200) PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key),
    calcutta_name VARCHAR(300) NOT NULL,
    budget_points INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bronze_calcuttas_tournament ON bronze_calcuttas(tournament_key);

-- Actual entry bids from real auction
CREATE TABLE bronze_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(200) NOT NULL REFERENCES bronze_calcuttas(calcutta_key),
    entry_key VARCHAR(200) NOT NULL,
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    bid_amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, entry_key, team_key)
);

CREATE INDEX idx_bronze_entry_bids_calcutta ON bronze_entry_bids(calcutta_key);
CREATE INDEX idx_bronze_entry_bids_entry ON bronze_entry_bids(entry_key);

-- Payout structure
CREATE TABLE bronze_payouts (
    id SERIAL PRIMARY KEY,
    calcutta_key VARCHAR(200) NOT NULL REFERENCES bronze_calcuttas(calcutta_key),
    position INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, position)
);

CREATE INDEX idx_bronze_payouts_calcutta ON bronze_payouts(calcutta_key);

-- ============================================================================
-- SILVER LAYER: Cleaned and enriched data
-- ============================================================================

-- Predicted game outcomes (ML model outputs)
CREATE TABLE silver_predicted_game_outcomes (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key),
    game_id VARCHAR(100) NOT NULL,
    round INTEGER NOT NULL,
    team1_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    team2_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    p_team1_wins DECIMAL(10,6) NOT NULL,
    p_matchup DECIMAL(10,6) NOT NULL DEFAULT 1.0,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_key, game_id)
);

CREATE INDEX idx_silver_pred_games_tournament ON silver_predicted_game_outcomes(tournament_key);

-- Predicted market share (ML model outputs)
CREATE TABLE silver_predicted_market_share (
    id BIGSERIAL PRIMARY KEY,
    calcutta_key VARCHAR(200) NOT NULL REFERENCES bronze_calcuttas(calcutta_key),
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    predicted_share_of_pool DECIMAL(10,6) NOT NULL,
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(calcutta_key, team_key)
);

CREATE INDEX idx_silver_pred_market_calcutta ON silver_predicted_market_share(calcutta_key);

-- Team tournament value (expected points)
CREATE TABLE silver_team_tournament_value (
    id BIGSERIAL PRIMARY KEY,
    tournament_key VARCHAR(100) NOT NULL REFERENCES bronze_tournaments(tournament_key),
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    expected_points DECIMAL(10,2) NOT NULL,
    p_champion DECIMAL(10,6),
    p_finals DECIMAL(10,6),
    p_final_four DECIMAL(10,6),
    p_elite_eight DECIMAL(10,6),
    p_sweet_sixteen DECIMAL(10,6),
    p_round_32 DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_key, team_key)
);

CREATE INDEX idx_silver_team_value_tournament ON silver_team_tournament_value(tournament_key);

-- ============================================================================
-- GOLD LAYER: Business metrics and analysis
-- ============================================================================

-- Optimization runs (tracks each strategy execution)
CREATE TABLE gold_optimization_runs (
    run_id VARCHAR(100) PRIMARY KEY,
    calcutta_key VARCHAR(200) NOT NULL REFERENCES bronze_calcuttas(calcutta_key),
    strategy VARCHAR(50) NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    budget_points INTEGER NOT NULL,
    run_timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gold_opt_runs_calcutta ON gold_optimization_runs(calcutta_key);
CREATE INDEX idx_gold_opt_runs_strategy ON gold_optimization_runs(strategy);

-- Recommended entry bids (optimizer outputs)
CREATE TABLE gold_recommended_entry_bids (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id),
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    bid_amount_points INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(run_id, team_key)
);

CREATE INDEX idx_gold_rec_bids_run ON gold_recommended_entry_bids(run_id);

-- Entry simulation outcomes (per simulation)
CREATE TABLE gold_entry_simulation_outcomes (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id),
    entry_key VARCHAR(200) NOT NULL,
    sim_id INTEGER NOT NULL,
    payout_cents INTEGER NOT NULL,
    total_points DECIMAL(10,2) NOT NULL,
    finish_position INTEGER NOT NULL,
    is_tied BOOLEAN NOT NULL,
    n_entries INTEGER NOT NULL,
    normalized_payout DECIMAL(10,6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(run_id, entry_key, sim_id)
);

CREATE INDEX idx_gold_entry_sim_run ON gold_entry_simulation_outcomes(run_id);
CREATE INDEX idx_gold_entry_sim_entry ON gold_entry_simulation_outcomes(entry_key);
CREATE INDEX idx_gold_entry_sim_run_entry ON gold_entry_simulation_outcomes(run_id, entry_key);

-- Entry performance summary (aggregated metrics)
CREATE TABLE gold_entry_performance (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id),
    entry_key VARCHAR(200) NOT NULL,
    is_our_strategy BOOLEAN NOT NULL DEFAULT FALSE,
    n_teams INTEGER NOT NULL,
    total_bid_points INTEGER NOT NULL,
    mean_payout_cents DECIMAL(10,2) NOT NULL,
    mean_points DECIMAL(10,2) NOT NULL,
    mean_normalized_payout DECIMAL(10,6) NOT NULL,
    p50_normalized_payout DECIMAL(10,6) NOT NULL,
    p90_normalized_payout DECIMAL(10,6) NOT NULL,
    p_top1 DECIMAL(10,6) NOT NULL,
    p_in_money DECIMAL(10,6) NOT NULL,
    percentile_rank DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(run_id, entry_key)
);

CREATE INDEX idx_gold_entry_perf_run ON gold_entry_performance(run_id);
CREATE INDEX idx_gold_entry_perf_rank ON gold_entry_performance(run_id, percentile_rank DESC);

-- Detailed investment report (team-level analysis for our strategy)
CREATE TABLE gold_detailed_investment_report (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(100) NOT NULL REFERENCES gold_optimization_runs(run_id),
    team_key VARCHAR(100) NOT NULL REFERENCES bronze_teams(team_key),
    our_bid_points INTEGER NOT NULL,
    expected_points DECIMAL(10,2) NOT NULL,
    predicted_market_points DECIMAL(10,2) NOT NULL,
    actual_market_points DECIMAL(10,2) NOT NULL,
    our_ownership DECIMAL(10,6) NOT NULL,
    expected_roi DECIMAL(10,4) NOT NULL,
    our_roi DECIMAL(10,4) NOT NULL,
    roi_degradation DECIMAL(10,4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(run_id, team_key)
);

CREATE INDEX idx_gold_detail_report_run ON gold_detailed_investment_report(run_id);

-- ============================================================================
-- VIEWS: Convenient query interfaces
-- ============================================================================

-- View: Latest optimization run per calcutta
CREATE VIEW view_latest_optimization_runs AS
SELECT DISTINCT ON (calcutta_key, strategy)
    run_id,
    calcutta_key,
    strategy,
    n_sims,
    seed,
    budget_points,
    run_timestamp
FROM gold_optimization_runs
ORDER BY calcutta_key, strategy, run_timestamp DESC;

-- View: Entry rankings with team details
CREATE VIEW view_entry_rankings AS
SELECT 
    ep.run_id,
    ep.entry_key,
    ep.is_our_strategy,
    ep.n_teams,
    ep.total_bid_points,
    ep.mean_normalized_payout,
    ep.percentile_rank,
    ep.p_top1,
    ep.p_in_money,
    RANK() OVER (PARTITION BY ep.run_id ORDER BY ep.mean_normalized_payout DESC) as rank,
    COUNT(*) OVER (PARTITION BY ep.run_id) as total_entries
FROM gold_entry_performance ep;

-- View: Tournament simulation statistics
CREATE VIEW view_tournament_sim_stats AS
SELECT 
    t.tournament_key,
    t.season,
    COUNT(DISTINCT st.sim_id) as n_sims,
    COUNT(DISTINCT st.team_key) as n_teams,
    AVG(st.wins + st.byes) as avg_progress,
    MAX(st.wins + st.byes) as max_progress
FROM bronze_tournaments t
JOIN bronze_simulated_tournaments st ON t.tournament_key = st.tournament_key
GROUP BY t.tournament_key, t.season;

-- ============================================================================
-- FUNCTIONS: Useful utilities
-- ============================================================================

-- Function to get entry portfolio
CREATE OR REPLACE FUNCTION get_entry_portfolio(p_run_id VARCHAR, p_entry_key VARCHAR)
RETURNS TABLE (
    team_key VARCHAR,
    school_name VARCHAR,
    seed INTEGER,
    region VARCHAR,
    bid_amount INTEGER
) AS $$
BEGIN
    IF p_entry_key = 'our_strategy' THEN
        RETURN QUERY
        SELECT 
            t.team_key,
            t.school_name,
            t.seed,
            t.region,
            reb.bid_amount_points as bid_amount
        FROM gold_recommended_entry_bids reb
        JOIN bronze_teams t ON reb.team_key = t.team_key
        WHERE reb.run_id = p_run_id
        ORDER BY reb.bid_amount_points DESC;
    ELSE
        RETURN QUERY
        SELECT 
            t.team_key,
            t.school_name,
            t.seed,
            t.region,
            eb.bid_amount
        FROM bronze_entry_bids eb
        JOIN bronze_teams t ON eb.team_key = t.team_key
        JOIN gold_optimization_runs r ON eb.calcutta_key = r.calcutta_key
        WHERE r.run_id = p_run_id AND eb.entry_key = p_entry_key
        ORDER BY eb.bid_amount DESC;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS: Documentation
-- ============================================================================

COMMENT ON SCHEMA public IS 'Calcutta Analytics - Medallion Architecture';
COMMENT ON TABLE bronze_tournaments IS 'Bronze: Raw tournament metadata';
COMMENT ON TABLE bronze_simulated_tournaments IS 'Bronze: Raw simulation outputs (compute-intensive, generated by Go service)';
COMMENT ON TABLE silver_predicted_game_outcomes IS 'Silver: ML predictions from Python sklearn models';
COMMENT ON TABLE gold_optimization_runs IS 'Gold: Optimization strategy executions';
COMMENT ON TABLE gold_entry_simulation_outcomes IS 'Gold: Per-simulation entry performance (for drill-down analysis)';
COMMENT ON TABLE gold_entry_performance IS 'Gold: Aggregated entry performance metrics';
