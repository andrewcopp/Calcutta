CREATE SCHEMA IF NOT EXISTS derived;

-- Drop compatibility medallion schemas (they only contain views).
DROP SCHEMA IF EXISTS bronze CASCADE;
DROP SCHEMA IF EXISTS silver CASCADE;
DROP SCHEMA IF EXISTS gold CASCADE;

-- Move lab-tier tables into derived (dropping lab/medallion naming).
ALTER TABLE IF EXISTS lab_bronze.tournaments SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_bronze.teams SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_bronze.calcuttas SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_bronze.entry_bids SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_bronze.payouts SET SCHEMA derived;

ALTER TABLE IF EXISTS lab_silver.predicted_game_outcomes SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_silver.predicted_market_share SET SCHEMA derived;

ALTER TABLE IF EXISTS lab_gold.strategy_generation_runs SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_gold.optimization_runs SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_gold.recommended_entry_bids SET SCHEMA derived;
ALTER TABLE IF EXISTS lab_gold.detailed_investment_report SET SCHEMA derived;

-- Drop the old lab schemas.
DROP SCHEMA IF EXISTS lab_bronze CASCADE;
DROP SCHEMA IF EXISTS lab_silver CASCADE;
DROP SCHEMA IF EXISTS lab_gold CASCADE;

-- analytics is now obsolete (tables moved to derived).
DROP SCHEMA IF EXISTS analytics CASCADE;
