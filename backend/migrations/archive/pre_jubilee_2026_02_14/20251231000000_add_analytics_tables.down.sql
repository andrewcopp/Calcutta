-- Drop analytics tables in reverse order

-- Drop functions
DROP FUNCTION IF EXISTS get_entry_portfolio(VARCHAR, VARCHAR);

-- Drop views
DROP VIEW IF EXISTS view_tournament_sim_stats;
DROP VIEW IF EXISTS view_entry_rankings;
DROP VIEW IF EXISTS view_latest_optimization_runs;

-- Drop gold layer tables
DROP TABLE IF EXISTS gold_detailed_investment_report;
DROP TABLE IF EXISTS gold_entry_performance;
DROP TABLE IF EXISTS gold_entry_simulation_outcomes;
DROP TABLE IF EXISTS gold_recommended_entry_bids;
DROP TABLE IF EXISTS gold_optimization_runs;

-- Drop silver layer tables
DROP TABLE IF EXISTS silver_team_tournament_value;
DROP TABLE IF EXISTS silver_predicted_market_share;
DROP TABLE IF EXISTS silver_predicted_game_outcomes;

-- Drop bronze layer tables
DROP TABLE IF EXISTS bronze_payouts;
DROP TABLE IF EXISTS bronze_entry_bids;
DROP TABLE IF EXISTS bronze_calcuttas;
DROP TABLE IF EXISTS bronze_simulated_tournaments;
DROP TABLE IF EXISTS bronze_teams;
DROP TABLE IF EXISTS bronze_tournaments;
