-- Rollback lab pipeline tables

-- Remove state column and constraint from lab.entries
DROP INDEX IF EXISTS idx_lab_entries_state;
ALTER TABLE lab.entries DROP CONSTRAINT IF EXISTS ck_lab_entries_state;
ALTER TABLE lab.entries DROP COLUMN IF EXISTS state;

-- Drop pipeline_calcutta_runs table
DROP TRIGGER IF EXISTS set_updated_at ON lab.pipeline_calcutta_runs;
DROP INDEX IF EXISTS uq_lab_pipeline_calcutta_runs_pipeline_calcutta;
DROP INDEX IF EXISTS idx_lab_pipeline_calcutta_runs_status;
DROP INDEX IF EXISTS idx_lab_pipeline_calcutta_runs_calcutta_id;
DROP INDEX IF EXISTS idx_lab_pipeline_calcutta_runs_pipeline_run_id;
DROP TABLE IF EXISTS lab.pipeline_calcutta_runs;

-- Drop pipeline_runs table
DROP TRIGGER IF EXISTS set_updated_at ON lab.pipeline_runs;
DROP INDEX IF EXISTS idx_lab_pipeline_runs_created_at;
DROP INDEX IF EXISTS idx_lab_pipeline_runs_status;
DROP INDEX IF EXISTS idx_lab_pipeline_runs_investment_model_id;
DROP TABLE IF EXISTS lab.pipeline_runs;
