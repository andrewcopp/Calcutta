-- Add human-readable names for strategy generation runs.

ALTER TABLE lab_gold.strategy_generation_runs
ADD COLUMN IF NOT EXISTS name TEXT;

UPDATE lab_gold.strategy_generation_runs
SET name = COALESCE(
    NULLIF(name, ''),
    NULLIF(params_json->>'name', ''),
    NULLIF(optimizer_key, ''),
    NULLIF(run_key, ''),
    'legacy'
)
WHERE name IS NULL OR name = '';
