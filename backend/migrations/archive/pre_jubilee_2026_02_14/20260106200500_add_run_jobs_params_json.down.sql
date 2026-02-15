CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_suite_calcutta_evaluation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        claimed_at,
        claimed_by,
        started_at,
        finished_at,
        error_message,
        created_at,
        updated_at
    )
    VALUES (
        'simulation',
        NEW.id,
        NEW.run_key,
        NEW.status,
        NEW.claimed_at,
        NEW.claimed_by,
        NEW.claimed_at,
        CASE WHEN NEW.status IN ('succeeded', 'failed') THEN NEW.updated_at ELSE NULL END,
        NEW.error_message,
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

ALTER TABLE IF EXISTS derived.run_jobs
    DROP COLUMN IF EXISTS params_json;
