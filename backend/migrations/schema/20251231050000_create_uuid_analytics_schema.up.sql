-- Create analytics tables with UUID primary keys for consistency with main schema
-- This allows eventual merging with legacy tables

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- BRONZE LAYER: Raw data from external sources
CREATE TABLE bronze_tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season INTEGER NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bronze_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    school_slug VARCHAR(255) NOT NULL,
    school_name VARCHAR(255) NOT NULL,
    seed INTEGER,
    region VARCHAR(50),
    kenpom_net DOUBLE PRECISION,
    kenpom_adj_em DOUBLE PRECISION,
    kenpom_adj_o DOUBLE PRECISION,
    kenpom_adj_d DOUBLE PRECISION,
    kenpom_adj_t DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tournament_id, school_slug)
);

CREATE TABLE bronze_calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    min_teams INTEGER NOT NULL DEFAULT 3,
    max_teams INTEGER NOT NULL DEFAULT 10,
    max_bid_points INTEGER NOT NULL DEFAULT 50,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bronze_entry_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    entry_name VARCHAR(255) NOT NULL,
    team_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    bid_amount_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bronze_payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    round INTEGER NOT NULL,
    points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SILVER LAYER: ML predictions and simulations
CREATE TABLE silver_predicted_game_outcomes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    game_id VARCHAR(255) NOT NULL,
    round INTEGER NOT NULL,
    team1_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    team2_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    p_team1_wins DOUBLE PRECISION NOT NULL,
    p_matchup DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    model_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tournament_id, game_id)
);

CREATE TABLE silver_simulated_tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES bronze_tournaments(id) ON DELETE CASCADE,
    sim_id INTEGER NOT NULL,
    team_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    wins INTEGER NOT NULL,
    byes INTEGER NOT NULL DEFAULT 0,
    eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tournament_id, sim_id, team_id)
);

CREATE TABLE silver_predicted_market_share (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    predicted_share DOUBLE PRECISION NOT NULL,
    predicted_points DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- GOLD LAYER: Optimization results and recommendations
CREATE TABLE gold_optimization_runs (
    run_id VARCHAR(255) PRIMARY KEY,
    calcutta_id UUID REFERENCES bronze_calcuttas(id) ON DELETE CASCADE,
    strategy VARCHAR(100) NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    budget_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gold_recommended_entry_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id VARCHAR(255) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    recommended_bid_points INTEGER NOT NULL,
    expected_roi DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gold_entry_simulation_outcomes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id VARCHAR(255) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    entry_name VARCHAR(255) NOT NULL,
    sim_id INTEGER NOT NULL,
    payout_points INTEGER NOT NULL,
    rank INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gold_entry_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id VARCHAR(255) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    entry_name VARCHAR(255) NOT NULL,
    mean_payout DOUBLE PRECISION NOT NULL,
    median_payout DOUBLE PRECISION NOT NULL,
    p_top1 DOUBLE PRECISION NOT NULL,
    p_in_money DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gold_detailed_investment_report (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id VARCHAR(255) NOT NULL REFERENCES gold_optimization_runs(run_id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES bronze_teams(id) ON DELETE CASCADE,
    expected_points DOUBLE PRECISION NOT NULL,
    predicted_market_points DOUBLE PRECISION NOT NULL,
    actual_market_points DOUBLE PRECISION,
    our_bid_points INTEGER,
    expected_roi DOUBLE PRECISION NOT NULL,
    our_roi DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_bronze_teams_tournament_id ON bronze_teams(tournament_id);
CREATE INDEX idx_bronze_teams_school_slug ON bronze_teams(school_slug);
CREATE INDEX idx_bronze_calcuttas_tournament_id ON bronze_calcuttas(tournament_id);
CREATE INDEX idx_bronze_entry_bids_calcutta_id ON bronze_entry_bids(calcutta_id);
CREATE INDEX idx_bronze_entry_bids_team_id ON bronze_entry_bids(team_id);

CREATE INDEX idx_silver_predicted_game_outcomes_tournament_id ON silver_predicted_game_outcomes(tournament_id);
CREATE INDEX idx_silver_simulated_tournaments_tournament_id ON silver_simulated_tournaments(tournament_id);
CREATE INDEX idx_silver_simulated_tournaments_sim_id ON silver_simulated_tournaments(tournament_id, sim_id);
CREATE INDEX idx_silver_simulated_tournaments_team_id ON silver_simulated_tournaments(team_id);

CREATE INDEX idx_gold_recommended_entry_bids_run_id ON gold_recommended_entry_bids(run_id);
CREATE INDEX idx_gold_recommended_entry_bids_team_id ON gold_recommended_entry_bids(team_id);
CREATE INDEX idx_gold_entry_simulation_outcomes_run_id ON gold_entry_simulation_outcomes(run_id);
CREATE INDEX idx_gold_entry_performance_run_id ON gold_entry_performance(run_id);
CREATE INDEX idx_gold_detailed_investment_report_run_id ON gold_detailed_investment_report(run_id);
