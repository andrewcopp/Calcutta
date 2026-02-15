ALTER TABLE IF EXISTS derived.strategy_generation_runs
    ADD COLUMN IF NOT EXISTS market_share_run_id UUID REFERENCES derived.market_share_runs(id),
    ADD COLUMN IF NOT EXISTS game_outcome_run_id UUID REFERENCES derived.game_outcome_runs(id),
    ADD COLUMN IF NOT EXISTS excluded_entry_name TEXT,
    ADD COLUMN IF NOT EXISTS starting_state_key TEXT;

CREATE INDEX IF NOT EXISTS idx_derived_strategy_generation_runs_market_share_run_id
ON derived.strategy_generation_runs(market_share_run_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_strategy_generation_runs_game_outcome_run_id
ON derived.strategy_generation_runs(game_outcome_run_id)
WHERE deleted_at IS NULL;

UPDATE derived.strategy_generation_runs
SET market_share_run_id = NULLIF(params_json->>'market_share_run_id', '')::uuid
WHERE market_share_run_id IS NULL
  AND params_json ? 'market_share_run_id'
  AND NULLIF(params_json->>'market_share_run_id', '') IS NOT NULL
  AND (params_json->>'market_share_run_id') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

UPDATE derived.strategy_generation_runs
SET game_outcome_run_id = NULLIF(params_json->>'game_outcome_run_id', '')::uuid
WHERE game_outcome_run_id IS NULL
  AND params_json ? 'game_outcome_run_id'
  AND NULLIF(params_json->>'game_outcome_run_id', '') IS NOT NULL
  AND (params_json->>'game_outcome_run_id') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

UPDATE derived.strategy_generation_runs
SET excluded_entry_name = NULLIF(params_json->>'excluded_entry_name', '')
WHERE excluded_entry_name IS NULL
  AND params_json ? 'excluded_entry_name'
  AND NULLIF(params_json->>'excluded_entry_name', '') IS NOT NULL;

UPDATE derived.strategy_generation_runs
SET starting_state_key = NULLIF(params_json->>'starting_state_key', '')
WHERE starting_state_key IS NULL
  AND params_json ? 'starting_state_key'
  AND NULLIF(params_json->>'starting_state_key', '') IS NOT NULL;

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
            'git_sha', NEW.git_sha,
            'market_share_run_id', NEW.market_share_run_id,
            'game_outcome_run_id', NEW.game_outcome_run_id,
            'excluded_entry_name', NEW.excluded_entry_name,
            'starting_state_key', NEW.starting_state_key
        ) || COALESCE(NEW.params_json, '{}'::jsonb),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;
