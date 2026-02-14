ALTER TABLE IF EXISTS derived.run_jobs
    ADD COLUMN IF NOT EXISTS progress_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE IF EXISTS derived.run_jobs
    ADD COLUMN IF NOT EXISTS progress_updated_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_progress_updated_at
ON derived.run_jobs(progress_updated_at);

UPDATE derived.run_jobs
SET progress_json = '{}'::jsonb
WHERE progress_json IS NULL;
