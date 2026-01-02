CREATE SCHEMA IF NOT EXISTS bronze;
CREATE SCHEMA IF NOT EXISTS silver;
CREATE SCHEMA IF NOT EXISTS gold;

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
