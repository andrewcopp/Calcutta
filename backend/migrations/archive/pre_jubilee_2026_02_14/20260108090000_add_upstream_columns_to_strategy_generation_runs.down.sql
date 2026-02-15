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

DROP INDEX IF EXISTS idx_derived_strategy_generation_runs_market_share_run_id;
DROP INDEX IF EXISTS idx_derived_strategy_generation_runs_game_outcome_run_id;

ALTER TABLE IF EXISTS derived.strategy_generation_runs
    DROP COLUMN IF EXISTS market_share_run_id,
    DROP COLUMN IF EXISTS game_outcome_run_id,
    DROP COLUMN IF EXISTS excluded_entry_name,
    DROP COLUMN IF EXISTS starting_state_key;
