-- Rollback: add_run_jobs_priority_dedup
-- Created: 2026-02-23 17:37:16 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Restore the original index
DROP INDEX IF EXISTS derived.idx_derived_run_jobs_claim;
CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_kind_status_created_at
    ON derived.run_jobs (run_kind, status, created_at);

-- Drop the dedup partial unique index
DROP INDEX IF EXISTS derived.uq_derived_run_jobs_dedup_key_active;

-- Drop new columns
ALTER TABLE derived.run_jobs
    DROP COLUMN IF EXISTS retry_after,
    DROP COLUMN IF EXISTS dedup_key,
    DROP COLUMN IF EXISTS priority;
