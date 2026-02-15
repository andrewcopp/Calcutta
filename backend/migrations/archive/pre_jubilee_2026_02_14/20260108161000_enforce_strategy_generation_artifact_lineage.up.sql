ALTER TABLE IF EXISTS derived.run_artifacts
    ADD CONSTRAINT ck_derived_run_artifacts_strategy_generation_lineage
    CHECK (
        run_kind <> 'strategy_generation'
        OR artifact_kind <> 'metrics'
        OR ((input_market_share_artifact_id IS NOT NULL) <> (input_advancement_artifact_id IS NOT NULL))
    );
