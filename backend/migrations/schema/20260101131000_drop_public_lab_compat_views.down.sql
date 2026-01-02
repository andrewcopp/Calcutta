-- Recreate public compatibility views for lab-tier tables.

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
