DROP SCHEMA IF EXISTS bronze CASCADE;
DROP SCHEMA IF EXISTS silver CASCADE;
DROP SCHEMA IF EXISTS gold CASCADE;

CREATE SCHEMA IF NOT EXISTS bronze;
CREATE SCHEMA IF NOT EXISTS silver;
CREATE SCHEMA IF NOT EXISTS gold;

-- Move product-facing derived tables back to their original medallion schemas
ALTER TABLE IF EXISTS analytics.simulated_tournaments SET SCHEMA silver;
ALTER TABLE IF EXISTS analytics.entry_simulation_outcomes SET SCHEMA gold;
ALTER TABLE IF EXISTS analytics.entry_performance SET SCHEMA gold;

-- Move lab-tier tables back to their original medallion schemas
ALTER TABLE IF EXISTS lab_bronze.tournaments SET SCHEMA bronze;
ALTER TABLE IF EXISTS lab_bronze.teams SET SCHEMA bronze;
ALTER TABLE IF EXISTS lab_bronze.calcuttas SET SCHEMA bronze;
ALTER TABLE IF EXISTS lab_bronze.entry_bids SET SCHEMA bronze;
ALTER TABLE IF EXISTS lab_bronze.payouts SET SCHEMA bronze;

ALTER TABLE IF EXISTS lab_silver.predicted_game_outcomes SET SCHEMA silver;
ALTER TABLE IF EXISTS lab_silver.predicted_market_share SET SCHEMA silver;

ALTER TABLE IF EXISTS lab_gold.optimization_runs SET SCHEMA gold;
ALTER TABLE IF EXISTS lab_gold.recommended_entry_bids SET SCHEMA gold;
ALTER TABLE IF EXISTS lab_gold.detailed_investment_report SET SCHEMA gold;

DROP SCHEMA IF EXISTS lab_bronze;
DROP SCHEMA IF EXISTS lab_silver;
DROP SCHEMA IF EXISTS lab_gold;

DROP SCHEMA IF EXISTS analytics;
