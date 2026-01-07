CREATE SCHEMA IF NOT EXISTS derived;

DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOR tbl IN
        SELECT quote_ident(schemaname) || '.' || quote_ident(tablename)
        FROM pg_tables
        WHERE schemaname = 'derived'
    LOOP
        EXECUTE 'TRUNCATE TABLE ' || tbl || ' CASCADE';
    END LOOP;
END $$;

DROP VIEW IF EXISTS derived.simulation_runs;
DROP VIEW IF EXISTS derived.simulation_run_batches;

DO $$
BEGIN
    IF to_regclass('derived.suites') IS NOT NULL THEN
        EXECUTE 'DROP TRIGGER IF EXISTS trg_sync_synthetic_calcutta_cohort_from_suite ON derived.suites';
    END IF;
    IF to_regclass('derived.suite_calcutta_evaluations') IS NOT NULL THEN
        EXECUTE 'DROP TRIGGER IF EXISTS trg_derived_suite_calcutta_evaluations_enqueue_run_job ON derived.suite_calcutta_evaluations';
    END IF;
END $$;

DROP FUNCTION IF EXISTS derived.sync_synthetic_calcutta_cohort_from_suite();
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_suite_calcutta_evaluation();

DROP TABLE IF EXISTS derived.synthetic_calcuttas;
DROP TABLE IF EXISTS derived.synthetic_calcutta_cohorts;
DROP TABLE IF EXISTS derived.suite_calcutta_evaluations;
DROP TABLE IF EXISTS derived.suite_executions;
DROP TABLE IF EXISTS derived.suite_scenarios;
DROP TABLE IF EXISTS derived.suites;
DROP TABLE IF EXISTS derived.run_artifacts;
DROP TABLE IF EXISTS derived.run_jobs;

CREATE TABLE IF NOT EXISTS derived.run_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_kind TEXT NOT NULL,
    run_id UUID NOT NULL,
    run_key UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    attempt INT NOT NULL DEFAULT 0,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    progress_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    progress_updated_at TIMESTAMPTZ,
    claimed_at TIMESTAMPTZ,
    claimed_by TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_derived_run_jobs_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_run_jobs_kind_run_id
ON derived.run_jobs(run_kind, run_id);

CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_kind_status_created_at
ON derived.run_jobs(run_kind, status, created_at);

CREATE INDEX IF NOT EXISTS idx_derived_run_jobs_progress_updated_at
ON derived.run_jobs(progress_updated_at);

DROP TRIGGER IF EXISTS set_updated_at ON derived.run_jobs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.run_jobs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.run_artifacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_kind TEXT NOT NULL,
    run_id UUID NOT NULL,
    run_key UUID,
    artifact_kind TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    storage_uri TEXT,
    summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_run_artifacts_kind_run_artifact
ON derived.run_artifacts(run_kind, run_id, artifact_kind)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_run_artifacts_run_key
ON derived.run_artifacts(run_key)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_run_artifacts_kind_run_id
ON derived.run_artifacts(run_kind, run_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.run_artifacts;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.run_artifacts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.synthetic_calcutta_cohorts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    game_outcomes_algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    market_share_algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    optimizer_key TEXT NOT NULL,
    n_sims INT NOT NULL,
    seed INT NOT NULL,
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name TEXT,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_starting_state_key
        CHECK (starting_state_key IN ('post_first_four', 'current')),
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_seed
        CHECK (seed <> 0),
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_n_sims
        CHECK (n_sims > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_synthetic_calcutta_cohorts_name
ON derived.synthetic_calcutta_cohorts(name)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcutta_cohorts;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcutta_cohorts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

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

CREATE TABLE IF NOT EXISTS derived.simulation_run_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cohort_id UUID NOT NULL REFERENCES derived.synthetic_calcutta_cohorts(id),
    name TEXT,
    optimizer_key TEXT,
    n_sims INT,
    seed INT,
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name TEXT,
    status TEXT NOT NULL DEFAULT 'running',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_simulation_run_batches_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
    CONSTRAINT ck_derived_simulation_run_batches_starting_state_key
        CHECK (starting_state_key IN ('post_first_four', 'current'))
);

CREATE INDEX IF NOT EXISTS idx_derived_simulation_run_batches_cohort_id
ON derived.simulation_run_batches(cohort_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_run_batches_created_at
ON derived.simulation_run_batches(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulation_run_batches;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulation_run_batches
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.simulation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_key UUID NOT NULL DEFAULT uuid_generate_v4(),
    simulation_run_batch_id UUID REFERENCES derived.simulation_run_batches(id),
    synthetic_calcutta_id UUID NOT NULL REFERENCES derived.synthetic_calcuttas(id),
    cohort_id UUID NOT NULL REFERENCES derived.synthetic_calcutta_cohorts(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    game_outcome_run_id UUID REFERENCES derived.game_outcome_runs(id),
    market_share_run_id UUID REFERENCES derived.market_share_runs(id),
    strategy_generation_run_id UUID REFERENCES derived.strategy_generation_runs(id),
    calcutta_evaluation_run_id UUID REFERENCES derived.calcutta_evaluation_runs(id),
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name TEXT,
    optimizer_key TEXT,
    n_sims INT,
    seed INT,
    our_rank INT,
    our_mean_normalized_payout DOUBLE PRECISION,
    our_median_normalized_payout DOUBLE PRECISION,
    our_p_top1 DOUBLE PRECISION,
    our_p_in_money DOUBLE PRECISION,
    total_simulations INT,
    realized_finish_position INT,
    realized_is_tied BOOLEAN,
    realized_in_the_money BOOLEAN,
    realized_payout_cents INT,
    realized_total_points DOUBLE PRECISION,
    status TEXT NOT NULL DEFAULT 'queued',
    claimed_at TIMESTAMPTZ,
    claimed_by TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_simulation_runs_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
    CONSTRAINT ck_derived_simulation_runs_starting_state_key
        CHECK (starting_state_key IN ('post_first_four', 'current'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_simulation_runs_batch_synthetic_calcutta
ON derived.simulation_runs(simulation_run_batch_id, synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_run_key
ON derived.simulation_runs(run_key)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_cohort_id
ON derived.simulation_runs(cohort_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_synthetic_calcutta_id
ON derived.simulation_runs(synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_calcutta_id
ON derived.simulation_runs(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_created_at
ON derived.simulation_runs(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulation_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulation_runs
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

DROP TRIGGER IF EXISTS trg_derived_simulation_runs_enqueue_run_job ON derived.simulation_runs;
CREATE TRIGGER trg_derived_simulation_runs_enqueue_run_job
AFTER INSERT ON derived.simulation_runs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_job_for_simulation_run();
