-- Rollback: remove is_benchmark column from lab.investment_models

ALTER TABLE lab.investment_models
DROP COLUMN IF EXISTS is_benchmark;
