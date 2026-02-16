ALTER TABLE derived.simulation_runs ADD COLUMN IF NOT EXISTS market_share_run_id UUID;
ALTER TABLE derived.simulation_runs ADD COLUMN IF NOT EXISTS strategy_generation_run_id UUID;
ALTER TABLE derived.simulation_runs ADD COLUMN IF NOT EXISTS optimizer_key TEXT;
