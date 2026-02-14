DROP INDEX IF EXISTS uq_derived_candidates_lab_config;
DROP INDEX IF EXISTS idx_derived_candidates_advancement_run_id;
DROP INDEX IF EXISTS idx_derived_candidates_market_share_artifact_id;
DROP INDEX IF EXISTS idx_derived_candidates_strategy_generation_run_id;
DROP INDEX IF EXISTS idx_derived_candidates_tournament_id;
DROP INDEX IF EXISTS idx_derived_candidates_calcutta_id;

ALTER TABLE IF EXISTS derived.candidates
    DROP COLUMN IF EXISTS git_sha,
    DROP COLUMN IF EXISTS excluded_entry_name,
    DROP COLUMN IF EXISTS starting_state_key,
    DROP COLUMN IF EXISTS optimizer_key,
    DROP COLUMN IF EXISTS advancement_run_id,
    DROP COLUMN IF EXISTS market_share_artifact_id,
    DROP COLUMN IF EXISTS market_share_run_id,
    DROP COLUMN IF EXISTS strategy_generation_run_id,
    DROP COLUMN IF EXISTS tournament_id,
    DROP COLUMN IF EXISTS calcutta_id;
