-- Recreate lab schemas and move tables back out of derived.

CREATE SCHEMA IF NOT EXISTS lab_bronze;
CREATE SCHEMA IF NOT EXISTS lab_silver;
CREATE SCHEMA IF NOT EXISTS lab_gold;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Recreate compatibility medallion schemas (views only).
CREATE SCHEMA IF NOT EXISTS bronze;
CREATE SCHEMA IF NOT EXISTS silver;
CREATE SCHEMA IF NOT EXISTS gold;

ALTER TABLE IF EXISTS derived.tournaments SET SCHEMA lab_bronze;

ALTER TABLE IF EXISTS derived.teams SET SCHEMA lab_bronze;

ALTER TABLE IF EXISTS derived.calcuttas SET SCHEMA lab_bronze;

ALTER TABLE IF EXISTS derived.entry_bids SET SCHEMA lab_bronze;

ALTER TABLE IF EXISTS derived.payouts SET SCHEMA lab_bronze;

ALTER TABLE IF EXISTS derived.predicted_game_outcomes SET SCHEMA lab_silver;

ALTER TABLE IF EXISTS derived.predicted_market_share SET SCHEMA lab_silver;

ALTER TABLE IF EXISTS derived.strategy_generation_runs SET SCHEMA lab_gold;

ALTER TABLE IF EXISTS derived.optimization_runs SET SCHEMA lab_gold;

ALTER TABLE IF EXISTS derived.recommended_entry_bids SET SCHEMA lab_gold;

ALTER TABLE IF EXISTS derived.detailed_investment_report SET SCHEMA lab_gold;

-- Restore compatibility views.
CREATE OR REPLACE VIEW bronze.tournaments AS SELECT * FROM lab_bronze.tournaments;
CREATE OR REPLACE VIEW bronze.teams AS SELECT * FROM lab_bronze.teams;
CREATE OR REPLACE VIEW bronze.calcuttas AS SELECT * FROM lab_bronze.calcuttas;
CREATE OR REPLACE VIEW bronze.entry_bids AS SELECT * FROM lab_bronze.entry_bids;
CREATE OR REPLACE VIEW bronze.payouts AS SELECT * FROM lab_bronze.payouts;

CREATE OR REPLACE VIEW silver.predicted_game_outcomes AS SELECT * FROM lab_silver.predicted_game_outcomes;
CREATE OR REPLACE VIEW silver.predicted_market_share AS SELECT * FROM lab_silver.predicted_market_share;

CREATE OR REPLACE VIEW gold.optimization_runs AS SELECT * FROM lab_gold.optimization_runs;
CREATE OR REPLACE VIEW gold.recommended_entry_bids AS SELECT * FROM lab_gold.recommended_entry_bids;
CREATE OR REPLACE VIEW gold.detailed_investment_report AS SELECT * FROM lab_gold.detailed_investment_report;
