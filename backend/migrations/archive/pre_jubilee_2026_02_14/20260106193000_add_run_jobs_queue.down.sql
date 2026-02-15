DROP TRIGGER IF EXISTS trg_derived_suite_calcutta_evaluations_enqueue_run_job ON derived.suite_calcutta_evaluations;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_suite_calcutta_evaluation();

DROP INDEX IF EXISTS idx_derived_suite_calcutta_evaluations_run_key;

ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    DROP COLUMN IF EXISTS run_key;

DROP TRIGGER IF EXISTS set_updated_at ON derived.run_jobs;

DROP INDEX IF EXISTS idx_derived_run_jobs_kind_status_created_at;
DROP INDEX IF EXISTS uq_derived_run_jobs_kind_run_id;

DROP TABLE IF EXISTS derived.run_jobs;
