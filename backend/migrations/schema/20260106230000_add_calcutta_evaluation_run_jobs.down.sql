DELETE FROM derived.run_jobs
WHERE run_kind = 'calcutta_evaluation';

DROP TRIGGER IF EXISTS trg_derived_calcutta_evaluation_runs_enqueue_run_job ON derived.calcutta_evaluation_runs;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_calcutta_evaluation_run();

DROP INDEX IF EXISTS idx_derived_calcutta_evaluation_runs_run_key;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    DROP COLUMN IF EXISTS git_sha;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    DROP COLUMN IF EXISTS params_json;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    DROP COLUMN IF EXISTS run_key;
