-- Rollback: Drop all UUID-based analytics tables

DROP TABLE IF EXISTS gold_detailed_investment_report CASCADE;
DROP TABLE IF EXISTS gold_entry_performance CASCADE;
DROP TABLE IF EXISTS gold_entry_simulation_outcomes CASCADE;
DROP TABLE IF EXISTS gold_recommended_entry_bids CASCADE;
DROP TABLE IF EXISTS gold_optimization_runs CASCADE;

DROP TABLE IF EXISTS silver_predicted_market_share CASCADE;
DROP TABLE IF EXISTS silver_simulated_tournaments CASCADE;
DROP TABLE IF EXISTS silver_predicted_game_outcomes CASCADE;

DROP TABLE IF EXISTS bronze_payouts CASCADE;
DROP TABLE IF EXISTS bronze_entry_bids CASCADE;
DROP TABLE IF EXISTS bronze_calcuttas CASCADE;
DROP TABLE IF EXISTS bronze_teams CASCADE;
DROP TABLE IF EXISTS bronze_tournaments CASCADE;
