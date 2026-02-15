DROP INDEX IF EXISTS idx_derived_run_jobs_progress_updated_at;

ALTER TABLE IF EXISTS derived.run_jobs
    DROP COLUMN IF EXISTS progress_updated_at;

ALTER TABLE IF EXISTS derived.run_jobs
    DROP COLUMN IF EXISTS progress_json;
