DELETE FROM derived.run_jobs
WHERE run_kind = 'strategy_generation';

DROP TRIGGER IF EXISTS trg_derived_strategy_generation_runs_enqueue_run_job ON derived.strategy_generation_runs;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_strategy_generation_run();

DROP INDEX IF EXISTS uq_derived_strategy_generation_runs_run_key_uuid;

ALTER TABLE IF EXISTS derived.strategy_generation_runs
    DROP COLUMN IF EXISTS run_key_uuid;
