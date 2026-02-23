-- Migration: add_run_jobs_priority_dedup
-- Created: 2026-02-23 17:37:16 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Add priority (lower = higher priority), dedup_key, and retry_after columns
ALTER TABLE derived.run_jobs
    ADD COLUMN IF NOT EXISTS priority smallint NOT NULL DEFAULT 10,
    ADD COLUMN IF NOT EXISTS dedup_key text,
    ADD COLUMN IF NOT EXISTS retry_after timestamptz;

-- Partial unique index: only one active job per dedup_key
CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_run_jobs_dedup_key_active
    ON derived.run_jobs (dedup_key)
    WHERE dedup_key IS NOT NULL AND status IN ('queued', 'running');

-- Replace the old kind+status+created_at index with a claim-optimized index
DROP INDEX IF EXISTS derived.idx_derived_run_jobs_kind_status_created_at;
CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_claim
    ON derived.run_jobs (status, priority, created_at);
