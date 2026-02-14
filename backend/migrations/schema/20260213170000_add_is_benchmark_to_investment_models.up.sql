-- Add is_benchmark flag to lab.investment_models
-- This distinguishes oracle/baseline models from real predictive models
-- for reporting and to prevent data leakage into training pipelines.

ALTER TABLE lab.investment_models
ADD COLUMN IF NOT EXISTS is_benchmark BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN lab.investment_models.is_benchmark IS
    'True for oracle/baseline models that should be excluded from training data';

-- Mark existing oracle model as benchmark
UPDATE lab.investment_models
SET is_benchmark = TRUE
WHERE kind = 'oracle' AND deleted_at IS NULL;
