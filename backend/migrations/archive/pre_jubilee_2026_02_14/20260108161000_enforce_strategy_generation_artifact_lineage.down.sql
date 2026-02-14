ALTER TABLE IF EXISTS derived.run_artifacts
    DROP CONSTRAINT IF EXISTS ck_derived_run_artifacts_strategy_generation_lineage;
