-- Restore orphaned simulation_cohorts and simulation_run_batches tables.

-- 1. Recreate simulation_cohorts
CREATE TABLE IF NOT EXISTS derived.simulation_cohorts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    game_outcomes_algorithm_id uuid NOT NULL,
    market_share_algorithm_id uuid NOT NULL,
    optimizer_key text NOT NULL,
    n_sims integer NOT NULL DEFAULT 10000,
    seed integer NOT NULL DEFAULT 42,
    starting_state_key text NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name text,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz
);

ALTER TABLE ONLY derived.simulation_cohorts
    ADD CONSTRAINT synthetic_calcutta_cohorts_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX uq_derived_synthetic_calcutta_cohorts_name
    ON derived.simulation_cohorts USING btree (name) WHERE (deleted_at IS NULL);

CREATE TRIGGER trg_derived_simulation_cohorts_updated_at
    BEFORE UPDATE ON derived.simulation_cohorts
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 2. Recreate simulation_run_batches
CREATE TABLE IF NOT EXISTS derived.simulation_run_batches (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    cohort_id uuid NOT NULL,
    name text,
    optimizer_key text,
    n_sims integer,
    seed integer,
    starting_state_key text NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name text,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT ck_derived_simulation_run_batches_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text]))),
    CONSTRAINT ck_derived_simulation_run_batches_status CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

ALTER TABLE ONLY derived.simulation_run_batches
    ADD CONSTRAINT simulation_run_batches_pkey PRIMARY KEY (id);

CREATE INDEX idx_derived_simulation_run_batches_cohort_id
    ON derived.simulation_run_batches USING btree (cohort_id) WHERE (deleted_at IS NULL);

CREATE INDEX idx_derived_simulation_run_batches_created_at
    ON derived.simulation_run_batches USING btree (created_at DESC) WHERE (deleted_at IS NULL);

CREATE TRIGGER trg_derived_simulation_run_batches_updated_at
    BEFORE UPDATE ON derived.simulation_run_batches
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 3. Restore columns on simulation_runs
ALTER TABLE derived.simulation_runs ADD COLUMN IF NOT EXISTS cohort_id uuid;
ALTER TABLE derived.simulation_runs ADD COLUMN IF NOT EXISTS simulation_run_batch_id uuid;

-- 4. Restore FKs
ALTER TABLE derived.simulation_run_batches
    ADD CONSTRAINT simulation_run_batches_cohort_id_fkey
    FOREIGN KEY (cohort_id) REFERENCES derived.simulation_cohorts(id);

ALTER TABLE derived.simulation_runs
    ADD CONSTRAINT simulation_runs_cohort_id_fkey
    FOREIGN KEY (cohort_id) REFERENCES derived.simulation_cohorts(id);

ALTER TABLE derived.simulation_runs
    ADD CONSTRAINT simulation_runs_simulation_run_batch_id_fkey
    FOREIGN KEY (simulation_run_batch_id) REFERENCES derived.simulation_run_batches(id);

-- 5. Restore trigger function with cohort_id and simulation_run_batch_id refs
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
