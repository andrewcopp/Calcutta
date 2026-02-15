-- Rollback strategy generation run tracking.

DROP INDEX IF EXISTS idx_lab_gold_strategy_generation_runs_created_at;
DROP INDEX IF EXISTS idx_lab_gold_strategy_generation_runs_calcutta_id;
DROP INDEX IF EXISTS idx_lab_gold_strategy_generation_runs_batch_id;

DROP TABLE IF EXISTS lab_gold.strategy_generation_runs CASCADE;
