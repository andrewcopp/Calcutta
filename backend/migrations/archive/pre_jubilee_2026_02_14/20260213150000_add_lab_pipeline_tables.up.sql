-- Lab Pipeline Tables: Orchestrate predict → optimize → evaluate workflow
-- Enables running full pipeline for investment models across historical calcuttas

--------------------------------------------------------------------------------
-- pipeline_runs: Orchestrates full pipeline for a model
--------------------------------------------------------------------------------
CREATE TABLE lab.pipeline_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investment_model_id UUID NOT NULL REFERENCES lab.investment_models(id),
    target_calcutta_ids UUID[] NOT NULL,
    budget_points INT NOT NULL DEFAULT 10000,
    optimizer_kind TEXT NOT NULL DEFAULT 'predicted_market_share',
    n_sims INT NOT NULL DEFAULT 10000,
    seed INT NOT NULL DEFAULT 42,
    excluded_entry_name TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_lab_pipeline_runs_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    CONSTRAINT ck_lab_pipeline_runs_budget_points CHECK (budget_points > 0),
    CONSTRAINT ck_lab_pipeline_runs_n_sims CHECK (n_sims > 0)
);

CREATE INDEX idx_lab_pipeline_runs_investment_model_id
ON lab.pipeline_runs(investment_model_id);

CREATE INDEX idx_lab_pipeline_runs_status
ON lab.pipeline_runs(status)
WHERE status IN ('pending', 'running');

CREATE INDEX idx_lab_pipeline_runs_created_at
ON lab.pipeline_runs(created_at DESC);

DROP TRIGGER IF EXISTS set_updated_at ON lab.pipeline_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab.pipeline_runs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

--------------------------------------------------------------------------------
-- pipeline_calcutta_runs: Per-calcutta tracking within a pipeline
--------------------------------------------------------------------------------
CREATE TABLE lab.pipeline_calcutta_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_run_id UUID NOT NULL REFERENCES lab.pipeline_runs(id) ON DELETE CASCADE,
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    entry_id UUID REFERENCES lab.entries(id),
    stage TEXT NOT NULL DEFAULT 'predictions',
    status TEXT NOT NULL DEFAULT 'pending',
    progress DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    progress_message TEXT,
    predictions_job_id UUID,
    optimization_job_id UUID,
    evaluation_job_id UUID,
    evaluation_id UUID REFERENCES lab.evaluations(id),
    error_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_lab_pipeline_calcutta_runs_stage
        CHECK (stage IN ('predictions', 'optimization', 'evaluation', 'completed')),
    CONSTRAINT ck_lab_pipeline_calcutta_runs_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed')),
    CONSTRAINT ck_lab_pipeline_calcutta_runs_progress
        CHECK (progress >= 0.0 AND progress <= 1.0)
);

CREATE INDEX idx_lab_pipeline_calcutta_runs_pipeline_run_id
ON lab.pipeline_calcutta_runs(pipeline_run_id);

CREATE INDEX idx_lab_pipeline_calcutta_runs_calcutta_id
ON lab.pipeline_calcutta_runs(calcutta_id);

CREATE INDEX idx_lab_pipeline_calcutta_runs_status
ON lab.pipeline_calcutta_runs(status)
WHERE status IN ('pending', 'running');

-- Unique constraint: one calcutta per pipeline run
CREATE UNIQUE INDEX uq_lab_pipeline_calcutta_runs_pipeline_calcutta
ON lab.pipeline_calcutta_runs(pipeline_run_id, calcutta_id);

DROP TRIGGER IF EXISTS set_updated_at ON lab.pipeline_calcutta_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab.pipeline_calcutta_runs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

--------------------------------------------------------------------------------
-- Add state column to lab.entries for tracking pipeline progress
--------------------------------------------------------------------------------
ALTER TABLE lab.entries ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'complete';

-- Add check constraint for valid states
ALTER TABLE lab.entries ADD CONSTRAINT ck_lab_entries_state
    CHECK (state IN ('pending_predictions', 'pending_optimization', 'pending_evaluation', 'complete'));

CREATE INDEX idx_lab_entries_state ON lab.entries(state)
WHERE state != 'complete';
