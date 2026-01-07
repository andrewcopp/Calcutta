ALTER TABLE IF EXISTS derived.strategy_generation_runs
    ADD COLUMN IF NOT EXISTS run_key_uuid UUID;

UPDATE derived.strategy_generation_runs
SET run_key_uuid = CASE
    WHEN run_key_uuid IS NOT NULL THEN run_key_uuid
    WHEN run_key IS NOT NULL AND run_key ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN run_key::uuid
    ELSE uuid_generate_v4()
END;

UPDATE derived.strategy_generation_runs
SET run_key = COALESCE(run_key, run_key_uuid::text)
WHERE run_key IS NULL;

ALTER TABLE IF EXISTS derived.strategy_generation_runs
    ALTER COLUMN run_key_uuid SET DEFAULT uuid_generate_v4();

ALTER TABLE derived.strategy_generation_runs
    ALTER COLUMN run_key_uuid SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_strategy_generation_runs_run_key_uuid
ON derived.strategy_generation_runs(run_key_uuid)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_strategy_generation_run()
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
        created_at,
        updated_at
    )
    VALUES (
        'strategy_generation',
        NEW.id,
        NEW.run_key_uuid,
        'queued',
        jsonb_build_object(
            'calcutta_id', NEW.calcutta_id,
            'simulated_tournament_id', NEW.simulated_tournament_id,
            'name', NEW.name,
            'purpose', NEW.purpose,
            'returns_model_key', NEW.returns_model_key,
            'investment_model_key', NEW.investment_model_key,
            'optimizer_key', NEW.optimizer_key,
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

DROP TRIGGER IF EXISTS trg_derived_strategy_generation_runs_enqueue_run_job ON derived.strategy_generation_runs;
CREATE TRIGGER trg_derived_strategy_generation_runs_enqueue_run_job
AFTER INSERT ON derived.strategy_generation_runs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_strategy_generation_run();

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
    'strategy_generation',
    r.id,
    r.run_key_uuid,
    'queued',
    0,
    jsonb_build_object(
        'calcutta_id', r.calcutta_id,
        'simulated_tournament_id', r.simulated_tournament_id,
        'name', r.name,
        'purpose', r.purpose,
        'returns_model_key', r.returns_model_key,
        'investment_model_key', r.investment_model_key,
        'optimizer_key', r.optimizer_key,
        'git_sha', r.git_sha
    ) || COALESCE(r.params_json, '{}'::jsonb),
    r.created_at,
    r.updated_at
FROM derived.strategy_generation_runs r
WHERE r.deleted_at IS NULL
ON CONFLICT (run_kind, run_id)
DO NOTHING;
