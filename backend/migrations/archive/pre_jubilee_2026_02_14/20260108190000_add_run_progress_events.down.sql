DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_update ON derived.run_jobs;
DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_insert ON derived.run_jobs;

DROP FUNCTION IF EXISTS derived.enqueue_run_progress_event_from_run_jobs();

DROP INDEX IF EXISTS idx_derived_run_progress_events_run_key;
DROP INDEX IF EXISTS idx_derived_run_progress_events_kind_run_created_at;

DROP TABLE IF EXISTS derived.run_progress_events;
