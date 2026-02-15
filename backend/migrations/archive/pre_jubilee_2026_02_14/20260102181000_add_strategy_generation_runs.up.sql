-- Add strategy generation run tracking (lab ML pipeline).

CREATE TABLE IF NOT EXISTS lab_gold.strategy_generation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_simulation_batch_id UUID REFERENCES analytics.tournament_simulation_batches(id),
    calcutta_id UUID REFERENCES core.calcuttas(id),
    purpose TEXT NOT NULL,
    returns_model_key TEXT NOT NULL,
    investment_model_key TEXT NOT NULL,
    optimizer_key TEXT NOT NULL,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    git_sha TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_lab_gold_strategy_generation_runs_batch_id
ON lab_gold.strategy_generation_runs(tournament_simulation_batch_id);

CREATE INDEX IF NOT EXISTS idx_lab_gold_strategy_generation_runs_calcutta_id
ON lab_gold.strategy_generation_runs(calcutta_id);

CREATE INDEX IF NOT EXISTS idx_lab_gold_strategy_generation_runs_created_at
ON lab_gold.strategy_generation_runs(created_at DESC);

DROP TRIGGER IF EXISTS set_updated_at ON lab_gold.strategy_generation_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab_gold.strategy_generation_runs
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();
