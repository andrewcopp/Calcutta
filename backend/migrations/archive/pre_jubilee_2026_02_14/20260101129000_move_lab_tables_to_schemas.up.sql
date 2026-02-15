-- Move lab-tier tables out of public into bronze/silver/gold schemas.
-- Keep backwards compatibility by creating updatable views in public with the old table names.

CREATE SCHEMA IF NOT EXISTS bronze;
CREATE SCHEMA IF NOT EXISTS silver;
CREATE SCHEMA IF NOT EXISTS gold;

-- BRONZE
ALTER TABLE IF EXISTS public.bronze_tournaments SET SCHEMA bronze;
ALTER TABLE IF EXISTS bronze.bronze_tournaments RENAME TO tournaments;

ALTER TABLE IF EXISTS public.bronze_teams SET SCHEMA bronze;
ALTER TABLE IF EXISTS bronze.bronze_teams RENAME TO teams;

ALTER TABLE IF EXISTS public.bronze_calcuttas SET SCHEMA bronze;
ALTER TABLE IF EXISTS bronze.bronze_calcuttas RENAME TO calcuttas;

ALTER TABLE IF EXISTS public.bronze_entry_bids SET SCHEMA bronze;
ALTER TABLE IF EXISTS bronze.bronze_entry_bids RENAME TO entry_bids;

ALTER TABLE IF EXISTS public.bronze_payouts SET SCHEMA bronze;
ALTER TABLE IF EXISTS bronze.bronze_payouts RENAME TO payouts;

-- SILVER
ALTER TABLE IF EXISTS public.silver_predicted_game_outcomes SET SCHEMA silver;
ALTER TABLE IF EXISTS silver.silver_predicted_game_outcomes RENAME TO predicted_game_outcomes;

ALTER TABLE IF EXISTS public.silver_simulated_tournaments SET SCHEMA silver;
ALTER TABLE IF EXISTS silver.silver_simulated_tournaments RENAME TO simulated_tournaments;

ALTER TABLE IF EXISTS public.silver_predicted_market_share SET SCHEMA silver;
ALTER TABLE IF EXISTS silver.silver_predicted_market_share RENAME TO predicted_market_share;

-- GOLD
ALTER TABLE IF EXISTS public.gold_optimization_runs SET SCHEMA gold;
ALTER TABLE IF EXISTS gold.gold_optimization_runs RENAME TO optimization_runs;

ALTER TABLE IF EXISTS public.gold_recommended_entry_bids SET SCHEMA gold;
ALTER TABLE IF EXISTS gold.gold_recommended_entry_bids RENAME TO recommended_entry_bids;

ALTER TABLE IF EXISTS public.gold_entry_simulation_outcomes SET SCHEMA gold;
ALTER TABLE IF EXISTS gold.gold_entry_simulation_outcomes RENAME TO entry_simulation_outcomes;

ALTER TABLE IF EXISTS public.gold_entry_performance SET SCHEMA gold;
ALTER TABLE IF EXISTS gold.gold_entry_performance RENAME TO entry_performance;

ALTER TABLE IF EXISTS public.gold_detailed_investment_report SET SCHEMA gold;
ALTER TABLE IF EXISTS gold.gold_detailed_investment_report RENAME TO detailed_investment_report;

-- Compatibility views (public)
CREATE OR REPLACE VIEW public.bronze_tournaments AS
SELECT * FROM bronze.tournaments;

CREATE OR REPLACE VIEW public.bronze_teams AS
SELECT * FROM bronze.teams;

CREATE OR REPLACE VIEW public.bronze_calcuttas AS
SELECT * FROM bronze.calcuttas;

CREATE OR REPLACE VIEW public.bronze_entry_bids AS
SELECT * FROM bronze.entry_bids;

CREATE OR REPLACE VIEW public.bronze_payouts AS
SELECT * FROM bronze.payouts;

CREATE OR REPLACE VIEW public.silver_predicted_game_outcomes AS
SELECT * FROM silver.predicted_game_outcomes;

CREATE OR REPLACE VIEW public.silver_simulated_tournaments AS
SELECT * FROM silver.simulated_tournaments;

CREATE OR REPLACE VIEW public.silver_predicted_market_share AS
SELECT * FROM silver.predicted_market_share;

CREATE OR REPLACE VIEW public.gold_optimization_runs AS
SELECT * FROM gold.optimization_runs;

CREATE OR REPLACE VIEW public.gold_recommended_entry_bids AS
SELECT * FROM gold.recommended_entry_bids;

CREATE OR REPLACE VIEW public.gold_entry_simulation_outcomes AS
SELECT * FROM gold.entry_simulation_outcomes;

CREATE OR REPLACE VIEW public.gold_entry_performance AS
SELECT * FROM gold.entry_performance;

CREATE OR REPLACE VIEW public.gold_detailed_investment_report AS
SELECT * FROM gold.detailed_investment_report;
