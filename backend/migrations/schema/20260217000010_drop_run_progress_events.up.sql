-- Drop triggers that write to run_progress_events from run_jobs
DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_insert ON derived.run_jobs;
DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_update ON derived.run_jobs;

-- Drop the trigger function
DROP FUNCTION IF EXISTS derived.enqueue_run_progress_event_from_run_jobs();

-- Drop the table
DROP TABLE IF EXISTS derived.run_progress_events;
