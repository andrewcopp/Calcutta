DELETE FROM derived.run_jobs
WHERE run_kind = 'market_share';

DROP TRIGGER IF EXISTS trg_derived_market_share_runs_enqueue_run_job ON derived.market_share_runs;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_market_share_run();

DROP INDEX IF EXISTS idx_derived_market_share_runs_run_key;

ALTER TABLE IF EXISTS derived.market_share_runs
    DROP COLUMN IF EXISTS run_key;
