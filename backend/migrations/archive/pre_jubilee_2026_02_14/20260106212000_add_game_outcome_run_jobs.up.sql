ALTER TABLE IF EXISTS derived.game_outcome_runs
    ADD COLUMN IF NOT EXISTS run_key UUID;

ALTER TABLE IF EXISTS derived.game_outcome_runs
    ALTER COLUMN run_key SET DEFAULT uuid_generate_v4();

UPDATE derived.game_outcome_runs
SET run_key = uuid_generate_v4()
WHERE run_key IS NULL;

ALTER TABLE derived.game_outcome_runs
    ALTER COLUMN run_key SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_run_key
ON derived.game_outcome_runs(run_key)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_game_outcome_run()
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
        'game_outcome',
        NEW.id,
        NEW.run_key,
        'queued',
        jsonb_build_object(
            'algorithm_id', NEW.algorithm_id,
            'tournament_id', NEW.tournament_id,
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

DROP TRIGGER IF EXISTS trg_derived_game_outcome_runs_enqueue_run_job ON derived.game_outcome_runs;
CREATE TRIGGER trg_derived_game_outcome_runs_enqueue_run_job
AFTER INSERT ON derived.game_outcome_runs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_game_outcome_run();

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
    'game_outcome',
    r.id,
    r.run_key,
    'queued',
    0,
    jsonb_build_object(
        'algorithm_id', r.algorithm_id,
        'tournament_id', r.tournament_id,
        'git_sha', r.git_sha
    ) || COALESCE(r.params_json, '{}'::jsonb),
    r.created_at,
    r.updated_at
FROM derived.game_outcome_runs r
WHERE r.deleted_at IS NULL
ON CONFLICT (run_kind, run_id)
DO NOTHING;
