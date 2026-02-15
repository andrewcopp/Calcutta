DROP INDEX IF EXISTS idx_derived_run_artifacts_input_market_share_artifact_id;
DROP INDEX IF EXISTS idx_derived_run_artifacts_input_advancement_artifact_id;

ALTER TABLE IF EXISTS derived.run_artifacts
    DROP COLUMN IF EXISTS input_market_share_artifact_id;

ALTER TABLE IF EXISTS derived.run_artifacts
    DROP COLUMN IF EXISTS input_advancement_artifact_id;
