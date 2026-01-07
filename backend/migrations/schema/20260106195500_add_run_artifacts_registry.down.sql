DROP TRIGGER IF EXISTS set_updated_at ON derived.run_artifacts;

DROP INDEX IF EXISTS idx_derived_run_artifacts_kind_run_id;
DROP INDEX IF EXISTS idx_derived_run_artifacts_run_key;
DROP INDEX IF EXISTS uq_derived_run_artifacts_kind_run_artifact;

DROP TABLE IF EXISTS derived.run_artifacts;
