ALTER TABLE IF EXISTS derived.run_jobs
    ADD COLUMN IF NOT EXISTS params_json JSONB NOT NULL DEFAULT '{}'::jsonb;

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

UPDATE derived.run_jobs j
SET params_json = jsonb_build_object(
    'optimizer_key', e.optimizer_key,
    'n_sims', e.n_sims,
    'seed', e.seed,
    'starting_state_key', e.starting_state_key,
    'excluded_entry_name', e.excluded_entry_name,
    'suite_id', e.suite_id,
    'suite_execution_id', e.suite_execution_id,
    'calcutta_id', e.calcutta_id,
    'game_outcome_run_id', e.game_outcome_run_id,
    'market_share_run_id', e.market_share_run_id,
    'strategy_generation_run_id', e.strategy_generation_run_id,
    'calcutta_evaluation_run_id', e.calcutta_evaluation_run_id
),
updated_at = NOW()
FROM derived.suite_calcutta_evaluations e
WHERE j.run_kind = 'simulation'
    AND j.run_id = e.id
    AND (j.params_json = '{}'::jsonb OR j.params_json IS NULL);
