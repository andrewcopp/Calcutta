CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS lab_bronze;
CREATE SCHEMA IF NOT EXISTS lab_silver;
CREATE SCHEMA IF NOT EXISTS lab_gold;

-- Move lab-tier tables into lab_* schemas (leaving bronze/silver/gold for compat views)
ALTER TABLE bronze.tournaments SET SCHEMA lab_bronze;
ALTER TABLE bronze.teams SET SCHEMA lab_bronze;
ALTER TABLE bronze.calcuttas SET SCHEMA lab_bronze;
ALTER TABLE bronze.entry_bids SET SCHEMA lab_bronze;
ALTER TABLE bronze.payouts SET SCHEMA lab_bronze;

ALTER TABLE silver.predicted_game_outcomes SET SCHEMA lab_silver;
ALTER TABLE silver.predicted_market_share SET SCHEMA lab_silver;

ALTER TABLE gold.optimization_runs SET SCHEMA lab_gold;
ALTER TABLE gold.recommended_entry_bids SET SCHEMA lab_gold;
ALTER TABLE gold.detailed_investment_report SET SCHEMA lab_gold;

-- Move product-facing derived tables into analytics
ALTER TABLE silver.simulated_tournaments SET SCHEMA analytics;
ALTER TABLE gold.entry_simulation_outcomes SET SCHEMA analytics;
ALTER TABLE gold.entry_performance SET SCHEMA analytics;

CREATE OR REPLACE VIEW bronze.tournaments AS SELECT * FROM lab_bronze.tournaments;
CREATE OR REPLACE VIEW bronze.teams AS SELECT * FROM lab_bronze.teams;
CREATE OR REPLACE VIEW bronze.calcuttas AS SELECT * FROM lab_bronze.calcuttas;
CREATE OR REPLACE VIEW bronze.entry_bids AS SELECT * FROM lab_bronze.entry_bids;
CREATE OR REPLACE VIEW bronze.payouts AS SELECT * FROM lab_bronze.payouts;

CREATE OR REPLACE VIEW silver.predicted_game_outcomes AS SELECT * FROM lab_silver.predicted_game_outcomes;
CREATE OR REPLACE VIEW silver.predicted_market_share AS SELECT * FROM lab_silver.predicted_market_share;
CREATE OR REPLACE VIEW silver.simulated_tournaments AS SELECT * FROM analytics.simulated_tournaments;

CREATE OR REPLACE VIEW gold.optimization_runs AS SELECT * FROM lab_gold.optimization_runs;
CREATE OR REPLACE VIEW gold.recommended_entry_bids AS SELECT * FROM lab_gold.recommended_entry_bids;
CREATE OR REPLACE VIEW gold.detailed_investment_report AS SELECT * FROM lab_gold.detailed_investment_report;
CREATE OR REPLACE VIEW gold.entry_simulation_outcomes AS SELECT * FROM analytics.entry_simulation_outcomes;
CREATE OR REPLACE VIEW gold.entry_performance AS SELECT * FROM analytics.entry_performance;
