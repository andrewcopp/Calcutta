-- Phase A5 (minimal) rollback: restore legacy synthetic calcuttas + candidates and synthetic_calcutta_id on simulation_runs.

ALTER TABLE IF EXISTS derived.simulation_runs
    ADD COLUMN IF NOT EXISTS synthetic_calcutta_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_simulation_runs_batch_synthetic_calcutta
ON derived.simulation_runs(simulation_run_batch_id, synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_synthetic_calcutta_id
ON derived.simulation_runs(synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS derived.synthetic_calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cohort_id UUID NOT NULL REFERENCES derived.synthetic_calcutta_cohorts(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    calcutta_snapshot_id UUID REFERENCES core.calcutta_snapshots(id),
    focus_strategy_generation_run_id UUID REFERENCES derived.strategy_generation_runs(id),
    focus_entry_name TEXT,
    starting_state_key TEXT,
    excluded_entry_name TEXT,
    highlighted_snapshot_entry_id UUID REFERENCES core.calcutta_snapshot_entries(id),
    notes TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_synthetic_calcuttas_starting_state_key
        CHECK (starting_state_key IS NULL OR starting_state_key IN ('post_first_four', 'current'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_synthetic_calcuttas_cohort_calcutta
ON derived.synthetic_calcuttas(cohort_id, calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_cohort_id
ON derived.synthetic_calcuttas(cohort_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_calcutta_id
ON derived.synthetic_calcuttas(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_created_at
ON derived.synthetic_calcuttas(created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_highlighted_snapshot_entry_id
ON derived.synthetic_calcuttas(highlighted_snapshot_entry_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcuttas;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcuttas
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.synthetic_calcutta_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    synthetic_calcutta_id UUID NOT NULL REFERENCES derived.synthetic_calcuttas(id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES derived.candidates(id) ON DELETE CASCADE,
    snapshot_entry_id UUID NOT NULL REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_synthetic_calcutta_candidates_synthetic_candidate
ON derived.synthetic_calcutta_candidates(synthetic_calcutta_id, candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_synthetic_calcutta_id
ON derived.synthetic_calcutta_candidates(synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_candidate_id
ON derived.synthetic_calcutta_candidates(candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_snapshot_entry_id
ON derived.synthetic_calcutta_candidates(snapshot_entry_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcutta_candidates;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcutta_candidates
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

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
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
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
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
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
