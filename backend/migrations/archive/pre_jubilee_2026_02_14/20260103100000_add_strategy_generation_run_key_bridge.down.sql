ALTER TABLE lab_gold.detailed_investment_report
DROP COLUMN IF EXISTS strategy_generation_run_id;

DROP INDEX IF EXISTS idx_lab_gold_detailed_investment_report_strategy_generation_run_id;

ALTER TABLE lab_gold.recommended_entry_bids
DROP COLUMN IF EXISTS strategy_generation_run_id;

DROP INDEX IF EXISTS idx_lab_gold_recommended_entry_bids_strategy_generation_run_id;

DROP INDEX IF EXISTS idx_lab_gold_optimization_runs_strategy_generation_run_id;

ALTER TABLE lab_gold.optimization_runs
DROP COLUMN IF EXISTS strategy_generation_run_id;

ALTER TABLE lab_gold.strategy_generation_runs
DROP CONSTRAINT IF EXISTS strategy_generation_runs_run_key_key;

ALTER TABLE lab_gold.strategy_generation_runs
DROP COLUMN IF EXISTS run_key;
