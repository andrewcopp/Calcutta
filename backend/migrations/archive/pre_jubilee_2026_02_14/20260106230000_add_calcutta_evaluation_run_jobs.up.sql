ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    ADD COLUMN IF NOT EXISTS run_key UUID;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    ALTER COLUMN run_key SET DEFAULT uuid_generate_v4();

UPDATE derived.calcutta_evaluation_runs
SET run_key = uuid_generate_v4()
WHERE run_key IS NULL;

ALTER TABLE derived.calcutta_evaluation_runs
    ALTER COLUMN run_key SET NOT NULL;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    ADD COLUMN IF NOT EXISTS params_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    ADD COLUMN IF NOT EXISTS git_sha TEXT;

CREATE INDEX IF NOT EXISTS idx_derived_calcutta_evaluation_runs_run_key
ON derived.calcutta_evaluation_runs(run_key)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF COALESCE(NEW.params_json->>'workerize', 'false') <> 'true' THEN
        RETURN NEW;
    END IF;

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        created_at,
        updated_at
    )
    VALUES (
        'calcutta_evaluation',
        NEW.id,
        NEW.run_key,
        'queued',
        jsonb_build_object(
            'simulated_tournament_id', NEW.simulated_tournament_id,
            'calcutta_snapshot_id', NEW.calcutta_snapshot_id,
            'purpose', NEW.purpose,
            'git_sha', NEW.git_sha
        ) || COALESCE(NEW.params_json, '{}'::jsonb),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_derived_calcutta_evaluation_runs_enqueue_run_job ON derived.calcutta_evaluation_runs;
CREATE TRIGGER trg_derived_calcutta_evaluation_runs_enqueue_run_job
AFTER INSERT ON derived.calcutta_evaluation_runs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run();

INSERT INTO derived.run_jobs (
    run_kind,
    run_id,
    run_key,
    status,
    attempt,
    params_json,
    created_at,
    updated_at
)
SELECT
    'calcutta_evaluation',
    r.id,
    r.run_key,
    'queued',
    0,
    jsonb_build_object(
        'simulated_tournament_id', r.simulated_tournament_id,
        'calcutta_snapshot_id', r.calcutta_snapshot_id,
        'purpose', r.purpose,
        'git_sha', r.git_sha
    ) || COALESCE(r.params_json, '{}'::jsonb),
    r.created_at,
    r.updated_at
FROM derived.calcutta_evaluation_runs r
WHERE r.deleted_at IS NULL
    AND COALESCE(r.params_json->>'workerize', 'false') = 'true'
ON CONFLICT (run_kind, run_id)
DO NOTHING;
