-- Drop orphan columns whose FK targets were archived
ALTER TABLE derived.simulation_runs DROP COLUMN IF EXISTS market_share_run_id;
ALTER TABLE derived.simulation_runs DROP COLUMN IF EXISTS strategy_generation_run_id;
ALTER TABLE derived.simulation_runs DROP COLUMN IF EXISTS optimizer_key;
