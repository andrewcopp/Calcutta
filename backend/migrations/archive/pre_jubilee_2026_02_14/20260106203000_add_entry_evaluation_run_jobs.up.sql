ALTER TABLE IF EXISTS derived.entry_evaluation_requests
    ADD COLUMN IF NOT EXISTS run_key UUID;

ALTER TABLE IF EXISTS derived.entry_evaluation_requests
    ALTER COLUMN run_key SET DEFAULT uuid_generate_v4();

UPDATE derived.entry_evaluation_requests
SET run_key = uuid_generate_v4()
WHERE run_key IS NULL;

ALTER TABLE derived.entry_evaluation_requests
    ALTER COLUMN run_key SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_run_key
ON derived.entry_evaluation_requests(run_key)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_entry_evaluation_request()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        claimed_at,
        claimed_by,
        started_at,
        finished_at,
        error_message,
        created_at,
        updated_at
    )
    VALUES (
        'entry_evaluation',
        NEW.id,
        NEW.run_key,
        NEW.status,
        jsonb_build_object(
            'calcutta_id', NEW.calcutta_id,
            'entry_candidate_id', NEW.entry_candidate_id,
            'excluded_entry_name', NEW.excluded_entry_name,
            'starting_state_key', NEW.starting_state_key,
            'n_sims', NEW.n_sims,
            'seed', NEW.seed,
            'experiment_key', NEW.experiment_key,
            'request_source', NEW.request_source
        ),
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

DROP TRIGGER IF EXISTS trg_derived_entry_evaluation_requests_enqueue_run_job ON derived.entry_evaluation_requests;
CREATE TRIGGER trg_derived_entry_evaluation_requests_enqueue_run_job
AFTER INSERT ON derived.entry_evaluation_requests
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_entry_evaluation_request();

INSERT INTO derived.run_jobs (
    run_kind,
    run_id,
    run_key,
    status,
    attempt,
    params_json,
    claimed_at,
    claimed_by,
    started_at,
    finished_at,
    error_message,
    created_at,
    updated_at
)
SELECT
    'entry_evaluation',
    r.id,
    r.run_key,
    r.status,
    CASE WHEN r.claimed_at IS NULL THEN 0 ELSE 1 END,
    jsonb_build_object(
        'calcutta_id', r.calcutta_id,
        'entry_candidate_id', r.entry_candidate_id,
        'excluded_entry_name', r.excluded_entry_name,
        'starting_state_key', r.starting_state_key,
        'n_sims', r.n_sims,
        'seed', r.seed,
        'experiment_key', r.experiment_key,
        'request_source', r.request_source
    ),
    r.claimed_at,
    r.claimed_by,
    r.claimed_at,
    CASE WHEN r.status IN ('succeeded', 'failed') THEN r.updated_at ELSE NULL END,
    r.error_message,
    r.created_at,
    r.updated_at
FROM derived.entry_evaluation_requests r
WHERE r.deleted_at IS NULL
ON CONFLICT (run_kind, run_id)
DO NOTHING;
