-- Revert lab-tier table move: drop public compatibility views and move tables back into public
-- with their original bronze_/silver_/gold_ names.

-- Drop compatibility views
DROP VIEW IF EXISTS public.gold_detailed_investment_report;
DROP VIEW IF EXISTS public.gold_entry_performance;
DROP VIEW IF EXISTS public.gold_entry_simulation_outcomes;
DROP VIEW IF EXISTS public.gold_recommended_entry_bids;
DROP VIEW IF EXISTS public.gold_optimization_runs;

DROP VIEW IF EXISTS public.silver_predicted_market_share;
DROP VIEW IF EXISTS public.silver_simulated_tournaments;
DROP VIEW IF EXISTS public.silver_predicted_game_outcomes;

DROP VIEW IF EXISTS public.bronze_payouts;
DROP VIEW IF EXISTS public.bronze_entry_bids;
DROP VIEW IF EXISTS public.bronze_calcuttas;
DROP VIEW IF EXISTS public.bronze_teams;
DROP VIEW IF EXISTS public.bronze_tournaments;

-- GOLD
ALTER TABLE IF EXISTS gold.detailed_investment_report RENAME TO gold_detailed_investment_report;
ALTER TABLE IF EXISTS gold.gold_detailed_investment_report SET SCHEMA public;

ALTER TABLE IF EXISTS gold.entry_performance RENAME TO gold_entry_performance;
ALTER TABLE IF EXISTS gold.gold_entry_performance SET SCHEMA public;

ALTER TABLE IF EXISTS gold.entry_simulation_outcomes RENAME TO gold_entry_simulation_outcomes;
ALTER TABLE IF EXISTS gold.gold_entry_simulation_outcomes SET SCHEMA public;

ALTER TABLE IF EXISTS gold.recommended_entry_bids RENAME TO gold_recommended_entry_bids;
ALTER TABLE IF EXISTS gold.gold_recommended_entry_bids SET SCHEMA public;

ALTER TABLE IF EXISTS gold.optimization_runs RENAME TO gold_optimization_runs;
ALTER TABLE IF EXISTS gold.gold_optimization_runs SET SCHEMA public;

-- SILVER
ALTER TABLE IF EXISTS silver.predicted_market_share RENAME TO silver_predicted_market_share;
ALTER TABLE IF EXISTS silver.silver_predicted_market_share SET SCHEMA public;

ALTER TABLE IF EXISTS silver.simulated_tournaments RENAME TO silver_simulated_tournaments;
ALTER TABLE IF EXISTS silver.silver_simulated_tournaments SET SCHEMA public;

ALTER TABLE IF EXISTS silver.predicted_game_outcomes RENAME TO silver_predicted_game_outcomes;
ALTER TABLE IF EXISTS silver.silver_predicted_game_outcomes SET SCHEMA public;

-- BRONZE
ALTER TABLE IF EXISTS bronze.payouts RENAME TO bronze_payouts;
ALTER TABLE IF EXISTS bronze.bronze_payouts SET SCHEMA public;

ALTER TABLE IF EXISTS bronze.entry_bids RENAME TO bronze_entry_bids;
ALTER TABLE IF EXISTS bronze.bronze_entry_bids SET SCHEMA public;

ALTER TABLE IF EXISTS bronze.calcuttas RENAME TO bronze_calcuttas;
ALTER TABLE IF EXISTS bronze.bronze_calcuttas SET SCHEMA public;

ALTER TABLE IF EXISTS bronze.teams RENAME TO bronze_teams;
ALTER TABLE IF EXISTS bronze.bronze_teams SET SCHEMA public;

ALTER TABLE IF EXISTS bronze.tournaments RENAME TO bronze_tournaments;
ALTER TABLE IF EXISTS bronze.bronze_tournaments SET SCHEMA public;
