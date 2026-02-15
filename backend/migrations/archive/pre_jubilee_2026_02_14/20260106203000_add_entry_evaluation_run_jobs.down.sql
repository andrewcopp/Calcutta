DROP TRIGGER IF EXISTS trg_derived_entry_evaluation_requests_enqueue_run_job ON derived.entry_evaluation_requests;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_entry_evaluation_request();

DROP INDEX IF EXISTS idx_derived_entry_evaluation_requests_run_key;

ALTER TABLE IF EXISTS derived.entry_evaluation_requests
    DROP COLUMN IF EXISTS run_key;
