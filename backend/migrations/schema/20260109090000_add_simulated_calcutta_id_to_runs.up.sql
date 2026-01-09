ALTER TABLE IF EXISTS derived.simulation_runs
    ADD COLUMN IF NOT EXISTS simulated_calcutta_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_derived_simulation_runs_simulated_calcutta_id'
    ) THEN
        ALTER TABLE derived.simulation_runs
            ADD CONSTRAINT fk_derived_simulation_runs_simulated_calcutta_id
            FOREIGN KEY (simulated_calcutta_id)
            REFERENCES derived.simulated_calcuttas(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_simulated_calcutta_id
ON derived.simulation_runs(simulated_calcutta_id)
WHERE deleted_at IS NULL;

ALTER TABLE IF EXISTS derived.simulation_runs
    ALTER COLUMN calcutta_id DROP NOT NULL;

ALTER TABLE IF EXISTS derived.simulation_runs
    ALTER COLUMN synthetic_calcutta_id DROP NOT NULL;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    ADD COLUMN IF NOT EXISTS simulated_calcutta_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_derived_calcutta_evaluation_runs_simulated_calcutta_id'
    ) THEN
        ALTER TABLE derived.calcutta_evaluation_runs
            ADD CONSTRAINT fk_derived_calcutta_evaluation_runs_simulated_calcutta_id
            FOREIGN KEY (simulated_calcutta_id)
            REFERENCES derived.simulated_calcuttas(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_derived_calcutta_evaluation_runs_simulated_calcutta_id
ON derived.calcutta_evaluation_runs(simulated_calcutta_id)
WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_simulation_run()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'synthetic_calcutta_id', NEW.synthetic_calcutta_id,
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'market_share_run_id', NEW.market_share_run_id,
        'strategy_generation_run_id', NEW.strategy_generation_run_id,
        'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
    );

    base_params := jsonb_build_object(
        'source', 'db_trigger',
        'optimizer_key', NEW.optimizer_key,
        'n_sims', NEW.n_sims,
        'seed', NEW.seed,
        'starting_state_key', NEW.starting_state_key,
        'excluded_entry_name', NEW.excluded_entry_name,
        'synthetic_calcutta_id', NEW.synthetic_calcutta_id,
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'market_share_run_id', NEW.market_share_run_id,
        'strategy_generation_run_id', NEW.strategy_generation_run_id,
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
