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
