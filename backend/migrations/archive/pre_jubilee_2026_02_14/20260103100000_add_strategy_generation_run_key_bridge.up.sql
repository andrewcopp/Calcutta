ALTER TABLE lab_gold.strategy_generation_runs
ADD COLUMN IF NOT EXISTS run_key TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'strategy_generation_runs_run_key_key'
    ) THEN
        ALTER TABLE lab_gold.strategy_generation_runs
        ADD CONSTRAINT strategy_generation_runs_run_key_key UNIQUE (run_key);
    END IF;
END $$;

ALTER TABLE lab_gold.optimization_runs
ADD COLUMN IF NOT EXISTS strategy_generation_run_id UUID REFERENCES lab_gold.strategy_generation_runs(id);

CREATE INDEX IF NOT EXISTS idx_lab_gold_optimization_runs_strategy_generation_run_id
ON lab_gold.optimization_runs(strategy_generation_run_id);

INSERT INTO lab_gold.strategy_generation_runs (
    run_key,
    tournament_simulation_batch_id,
    calcutta_id,
    purpose,
    returns_model_key,
    investment_model_key,
    optimizer_key,
    params_json,
    git_sha,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    r.run_id AS run_key,
    NULL::uuid AS tournament_simulation_batch_id,
    bc.core_calcutta_id AS calcutta_id,
    'legacy_optimization_run' AS purpose,
    'legacy' AS returns_model_key,
    'legacy' AS investment_model_key,
    COALESCE(NULLIF(r.strategy, ''), 'legacy') AS optimizer_key,
    '{}'::jsonb AS params_json,
    NULL::text AS git_sha,
    r.created_at,
    r.updated_at,
    r.deleted_at
FROM lab_gold.optimization_runs r
LEFT JOIN lab_bronze.calcuttas bc ON bc.id = r.calcutta_id
WHERE r.run_id IS NOT NULL
ON CONFLICT (run_key) DO UPDATE SET
    updated_at = EXCLUDED.updated_at;

UPDATE lab_gold.optimization_runs r
SET strategy_generation_run_id = sgr.id
FROM lab_gold.strategy_generation_runs sgr
WHERE sgr.run_key = r.run_id
  AND r.strategy_generation_run_id IS NULL;

ALTER TABLE lab_gold.recommended_entry_bids
ADD COLUMN IF NOT EXISTS strategy_generation_run_id UUID REFERENCES lab_gold.strategy_generation_runs(id);

CREATE INDEX IF NOT EXISTS idx_lab_gold_recommended_entry_bids_strategy_generation_run_id
ON lab_gold.recommended_entry_bids(strategy_generation_run_id);

UPDATE lab_gold.recommended_entry_bids reb
SET strategy_generation_run_id = r.strategy_generation_run_id
FROM lab_gold.optimization_runs r
WHERE reb.run_id = r.run_id
  AND reb.strategy_generation_run_id IS NULL;

ALTER TABLE lab_gold.detailed_investment_report
ADD COLUMN IF NOT EXISTS strategy_generation_run_id UUID REFERENCES lab_gold.strategy_generation_runs(id);

CREATE INDEX IF NOT EXISTS idx_lab_gold_detailed_investment_report_strategy_generation_run_id
ON lab_gold.detailed_investment_report(strategy_generation_run_id);

UPDATE lab_gold.detailed_investment_report dir
SET strategy_generation_run_id = r.strategy_generation_run_id
FROM lab_gold.optimization_runs r
WHERE dir.run_id = r.run_id
  AND dir.strategy_generation_run_id IS NULL;
