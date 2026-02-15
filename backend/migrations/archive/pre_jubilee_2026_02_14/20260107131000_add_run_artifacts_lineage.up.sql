ALTER TABLE IF EXISTS derived.run_artifacts
    ADD COLUMN IF NOT EXISTS input_market_share_artifact_id UUID REFERENCES derived.run_artifacts(id);

ALTER TABLE IF EXISTS derived.run_artifacts
    ADD COLUMN IF NOT EXISTS input_advancement_artifact_id UUID REFERENCES derived.run_artifacts(id);

-- Backfill lineage for strategy_generation metrics artifacts using the originating run_jobs params_json.
UPDATE derived.run_artifacts a
SET input_market_share_artifact_id = ms.id
FROM derived.run_jobs j
JOIN derived.run_artifacts ms
    ON ms.run_kind = 'market_share'
    AND ms.run_id = (j.params_json->>'market_share_run_id')::uuid
    AND ms.artifact_kind = 'metrics'
    AND ms.deleted_at IS NULL
WHERE a.run_kind = 'strategy_generation'
    AND a.artifact_kind = 'metrics'
    AND a.deleted_at IS NULL
    AND a.input_market_share_artifact_id IS NULL
    AND j.run_kind = 'strategy_generation'
    AND j.run_id = a.run_id
    AND (j.params_json->>'market_share_run_id') IS NOT NULL
    AND (j.params_json->>'market_share_run_id') <> '';

CREATE INDEX IF NOT EXISTS idx_derived_run_artifacts_input_market_share_artifact_id
ON derived.run_artifacts(input_market_share_artifact_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_run_artifacts_input_advancement_artifact_id
ON derived.run_artifacts(input_advancement_artifact_id)
WHERE deleted_at IS NULL;
