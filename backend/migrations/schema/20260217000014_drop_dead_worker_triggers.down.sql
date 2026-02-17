-- Recreate trigger functions (final versions after fix migrations)

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    seed_from_sim INTEGER;
    seed_from_params INTEGER;
    seed_value INTEGER;
    dataset_refs JSONB;
    base_params JSONB;
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

    dataset_refs := jsonb_build_object(
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'calcutta_snapshot_id', NEW.calcutta_snapshot_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'seed', seed_value,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'calcutta_snapshot_id', NEW.calcutta_snapshot_id,
        'purpose', NEW.purpose,
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
        'calcutta_evaluation',
        NEW.id,
        NEW.run_key,
        'queued',
        ((base_params || COALESCE(NEW.params_json, '{}'::jsonb)) || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

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

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_simulation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
        'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
    );

    base_params := jsonb_build_object(
        'source', 'db_trigger',
        'n_sims', NEW.n_sims,
        'seed', NEW.seed,
        'starting_state_key', NEW.starting_state_key,
        'excluded_entry_name', NEW.excluded_entry_name,
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
        'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
    );

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
        (base_params || jsonb_build_object('dataset_refs', dataset_refs)),
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

-- Recreate triggers
CREATE TRIGGER trg_derived_calcutta_evaluation_runs_enqueue_run_job
    AFTER INSERT ON derived.calcutta_evaluation_runs
    FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run();

CREATE TRIGGER trg_derived_game_outcome_runs_enqueue_run_job
    AFTER INSERT ON derived.game_outcome_runs
    FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_game_outcome_run();

CREATE TRIGGER trg_derived_simulation_runs_enqueue_run_job
    AFTER INSERT ON derived.simulation_runs
    FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_simulation_run();
