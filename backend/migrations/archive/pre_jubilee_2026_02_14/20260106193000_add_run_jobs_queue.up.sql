CREATE TABLE IF NOT EXISTS derived.run_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_kind TEXT NOT NULL,
    run_id UUID NOT NULL,
    run_key UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    attempt INT NOT NULL DEFAULT 0,
    claimed_at TIMESTAMPTZ,
    claimed_by TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_derived_run_jobs_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_run_jobs_kind_run_id
ON derived.run_jobs(run_kind, run_id);

CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_kind_status_created_at
ON derived.run_jobs(run_kind, status, created_at);

DROP TRIGGER IF EXISTS set_updated_at ON derived.run_jobs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.run_jobs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    ADD COLUMN IF NOT EXISTS run_key UUID;

ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    ALTER COLUMN run_key SET DEFAULT uuid_generate_v4();

UPDATE derived.suite_calcutta_evaluations
SET run_key = uuid_generate_v4()
WHERE run_key IS NULL;

ALTER TABLE derived.suite_calcutta_evaluations
    ALTER COLUMN run_key SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_calcutta_evaluations_run_key
ON derived.suite_calcutta_evaluations(run_key)
WHERE deleted_at IS NULL;

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

DROP TRIGGER IF EXISTS trg_derived_suite_calcutta_evaluations_enqueue_run_job ON derived.suite_calcutta_evaluations;
CREATE TRIGGER trg_derived_suite_calcutta_evaluations_enqueue_run_job
AFTER INSERT ON derived.suite_calcutta_evaluations
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_suite_calcutta_evaluation();

INSERT INTO derived.run_jobs (
    run_kind,
    run_id,
    run_key,
    status,
    attempt,
    claimed_at,
    claimed_by,
    started_at,
    finished_at,
    error_message,
    created_at,
    updated_at
)
SELECT
    'simulation',
    e.id,
    e.run_key,
    e.status,
    CASE WHEN e.claimed_at IS NULL THEN 0 ELSE 1 END,
    e.claimed_at,
    e.claimed_by,
    e.claimed_at,
    CASE WHEN e.status IN ('succeeded', 'failed') THEN e.updated_at ELSE NULL END,
    e.error_message,
    e.created_at,
    e.updated_at
FROM derived.suite_calcutta_evaluations e
WHERE e.deleted_at IS NULL
ON CONFLICT (run_kind, run_id)
DO NOTHING;
