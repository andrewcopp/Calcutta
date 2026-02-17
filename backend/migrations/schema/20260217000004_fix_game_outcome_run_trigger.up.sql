-- Fix the enqueue_run_job_for_game_outcome_run trigger function.
-- The algorithm_id column was dropped from game_outcome_runs in migration
-- 20260217000003, but this trigger still referenced NEW.algorithm_id.

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_game_outcome_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'tournament_id', NEW.tournament_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'tournament_id', NEW.tournament_id,
        'git_sha', NEW.git_sha
    );

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
        base_params,
        NOW(),
        NOW()
    )
    ON CONFLICT (run_kind, run_id) DO UPDATE SET
        status = 'queued',
        params_json = EXCLUDED.params_json,
        updated_at = NOW();

    RETURN NEW;
END;
$$;
