CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_strategy_generation_run()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'calcutta_id', NEW.calcutta_id,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'market_share_run_id', NEW.market_share_run_id,
        'game_outcome_run_id', NEW.game_outcome_run_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'calcutta_id', NEW.calcutta_id,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'name', NEW.name,
        'purpose', NEW.purpose,
        'returns_model_key', NEW.returns_model_key,
        'investment_model_key', NEW.investment_model_key,
        'optimizer_key', NEW.optimizer_key,
        'git_sha', NEW.git_sha,
        'market_share_run_id', NEW.market_share_run_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'excluded_entry_name', NEW.excluded_entry_name,
        'starting_state_key', NEW.starting_state_key
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
        'strategy_generation',
        NEW.id,
        NEW.run_key_uuid,
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

UPDATE derived.run_jobs j
SET params_json = COALESCE(j.params_json, '{}'::jsonb) || jsonb_build_object(
    'dataset_refs', jsonb_build_object(
        'calcutta_id', r.calcutta_id,
        'simulated_tournament_id', r.simulated_tournament_id,
        'market_share_run_id', r.market_share_run_id,
        'game_outcome_run_id', r.game_outcome_run_id
    )
)
FROM derived.strategy_generation_runs r
WHERE j.run_kind = 'strategy_generation'
    AND j.run_id = r.id
    AND (j.params_json->'dataset_refs' IS NULL OR j.params_json->'dataset_refs' = 'null'::jsonb);
