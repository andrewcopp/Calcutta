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
        'simulation',
        NEW.id,
        NEW.run_key,
        NEW.status,
        jsonb_build_object(
            'source', 'db_trigger',
            'optimizer_key', NEW.optimizer_key,
            'n_sims', NEW.n_sims,
            'seed', NEW.seed,
            'starting_state_key', NEW.starting_state_key,
            'excluded_entry_name', NEW.excluded_entry_name,
            'suite_id', NEW.suite_id,
            'suite_execution_id', NEW.suite_execution_id,
            'calcutta_id', NEW.calcutta_id,
            'game_outcome_run_id', NEW.game_outcome_run_id,
            'market_share_run_id', NEW.market_share_run_id,
            'strategy_generation_run_id', NEW.strategy_generation_run_id,
            'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
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
            'source', COALESCE(NULLIF(NEW.request_source, ''), 'db_trigger'),
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

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_market_share_run()
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
        'market_share',
        NEW.id,
        NEW.run_key,
        'queued',
        jsonb_build_object(
            'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
            'algorithm_id', NEW.algorithm_id,
            'calcutta_id', NEW.calcutta_id,
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
            'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
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
            'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
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

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    seed_from_sim INTEGER;
    seed_from_params INTEGER;
    seed_value INTEGER;
BEGIN
    IF COALESCE(NEW.params_json->>'workerize', 'false') <> 'true' THEN
        RETURN NEW;
    END IF;

    seed_from_sim := NULL;
    SELECT st.seed
    INTO seed_from_sim
    FROM derived.simulated_tournaments st
    WHERE st.id = NEW.simulated_tournament_id
        AND st.deleted_at IS NULL
    LIMIT 1;

    seed_from_params := NULL;
    BEGIN
        seed_from_params := NULLIF(COALESCE(NEW.params_json->>'seed', ''), '')::int;
    EXCEPTION WHEN OTHERS THEN
        seed_from_params := NULL;
    END;

    seed_value := COALESCE(seed_from_params, seed_from_sim);

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
            'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
            'seed', seed_value,
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

UPDATE derived.run_jobs
SET params_json = COALESCE(params_json, '{}'::jsonb) - 'dataset_refs'
WHERE params_json ? 'dataset_refs';
