-- Drop orphaned simulation_cohorts and simulation_run_batches tables.
-- These tables are not referenced by any application code.

-- 1. Drop FKs that reference these tables
ALTER TABLE derived.simulation_runs DROP CONSTRAINT IF EXISTS simulation_runs_cohort_id_fkey;
ALTER TABLE derived.simulation_runs DROP CONSTRAINT IF EXISTS simulation_runs_simulation_run_batch_id_fkey;
ALTER TABLE derived.simulation_run_batches DROP CONSTRAINT IF EXISTS simulation_run_batches_cohort_id_fkey;

-- 2. Drop the orphaned columns from simulation_runs
ALTER TABLE derived.simulation_runs DROP COLUMN IF EXISTS cohort_id;
ALTER TABLE derived.simulation_runs DROP COLUMN IF EXISTS simulation_run_batch_id;

-- 3. Drop the orphaned tables
DROP TABLE IF EXISTS derived.simulation_run_batches;
DROP TABLE IF EXISTS derived.simulation_cohorts;

-- 4. Update the trigger function to remove references to dropped columns
CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_simulation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
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
