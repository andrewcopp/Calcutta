-- Recreate run_artifacts table
CREATE TABLE derived.run_artifacts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid,
    artifact_kind text NOT NULL,
    schema_version text NOT NULL DEFAULT 'v1'::text,
    storage_uri text,
    summary_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    input_market_share_artifact_id uuid,
    input_advancement_artifact_id uuid,
    CONSTRAINT ck_derived_run_artifacts_strategy_generation_lineage CHECK (((run_kind <> 'strategy_generation'::text) OR (artifact_kind <> 'metrics'::text) OR ((input_market_share_artifact_id IS NOT NULL) <> (input_advancement_artifact_id IS NOT NULL))))
);

ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_pkey PRIMARY KEY (id);
CREATE INDEX idx_derived_run_artifacts_input_advancement_artifact_id ON derived.run_artifacts USING btree (input_advancement_artifact_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_input_market_share_artifact_id ON derived.run_artifacts USING btree (input_market_share_artifact_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_kind_run_id ON derived.run_artifacts USING btree (run_kind, run_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_run_key ON derived.run_artifacts USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_run_artifacts_kind_run_artifact ON derived.run_artifacts USING btree (run_kind, run_id, artifact_kind) WHERE (deleted_at IS NULL);
ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_input_advancement_artifact_id_fkey FOREIGN KEY (input_advancement_artifact_id) REFERENCES derived.run_artifacts(id);
ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_input_market_share_artifact_id_fkey FOREIGN KEY (input_market_share_artifact_id) REFERENCES derived.run_artifacts(id);
CREATE TRIGGER trg_derived_run_artifacts_updated_at BEFORE UPDATE ON derived.run_artifacts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Recreate predicted_market_share table
CREATE TABLE derived.predicted_market_share (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid,
    team_id uuid NOT NULL,
    tournament_id uuid,
    predicted_market_share double precision NOT NULL,
    model_key text,
    run_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT market_share_must_have_calcutta_or_tournament CHECK (((calcutta_id IS NOT NULL) OR (tournament_id IS NOT NULL)))
);

ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT silver_predicted_market_share_pkey PRIMARY KEY (id);
CREATE INDEX idx_derived_predicted_market_share_calcutta_id ON derived.predicted_market_share USING btree (calcutta_id) WHERE ((calcutta_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_derived_predicted_market_share_run_id ON derived.predicted_market_share USING btree (run_id) WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_derived_predicted_market_share_tournament_id ON derived.predicted_market_share USING btree (tournament_id) WHERE ((tournament_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_derived_predicted_market_share_run_team ON derived.predicted_market_share USING btree (run_id, team_id) WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
CREATE TRIGGER trg_derived_predicted_market_share_updated_at BEFORE UPDATE ON derived.predicted_market_share FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
