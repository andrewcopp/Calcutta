-- Rollback strategy generation run human-readable names.

ALTER TABLE lab_gold.strategy_generation_runs
DROP COLUMN IF EXISTS name;
