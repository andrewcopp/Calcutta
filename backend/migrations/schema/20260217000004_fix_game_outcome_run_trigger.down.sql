-- Restore the original trigger function that references algorithm_id.
-- Only valid if algorithm_id column has been restored by 20260217000003 down.

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
        'algorithm_id', NEW.algorithm_id,
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
        'pending',
        base_params,
        NOW(),
        NOW()
    )
    ON CONFLICT (run_kind, run_id) DO UPDATE SET
        status = 'pending',
        params_json = EXCLUDED.params_json,
        updated_at = NOW();

    RETURN NEW;
END;
$$;
