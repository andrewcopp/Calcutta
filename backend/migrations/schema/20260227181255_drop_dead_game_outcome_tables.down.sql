-- Recreate compute.game_outcome_runs
CREATE TABLE compute.game_outcome_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    git_sha text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid DEFAULT public.uuid_generate_v4() NOT NULL
);

ALTER TABLE ONLY compute.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_pkey PRIMARY KEY (id);

CREATE INDEX idx_derived_game_outcome_runs_created_at
    ON compute.game_outcome_runs USING btree (created_at DESC)
    WHERE (deleted_at IS NULL);

CREATE INDEX idx_derived_game_outcome_runs_run_key
    ON compute.game_outcome_runs USING btree (run_key)
    WHERE (deleted_at IS NULL);

CREATE INDEX idx_derived_game_outcome_runs_tournament_id
    ON compute.game_outcome_runs USING btree (tournament_id)
    WHERE (deleted_at IS NULL);

CREATE TRIGGER trg_derived_game_outcome_runs_updated_at
    BEFORE UPDATE ON compute.game_outcome_runs
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE ONLY compute.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);

-- Recreate compute.predicted_game_outcomes
CREATE TABLE compute.predicted_game_outcomes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    game_id character varying(255) NOT NULL,
    round integer NOT NULL,
    team1_id uuid NOT NULL,
    team2_id uuid NOT NULL,
    p_team1_wins double precision NOT NULL,
    p_matchup double precision DEFAULT 1.0 NOT NULL,
    model_version character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_id uuid
);

ALTER TABLE ONLY compute.predicted_game_outcomes
    ADD CONSTRAINT silver_predicted_game_outcomes_pkey PRIMARY KEY (id);

CREATE INDEX idx_derived_predicted_game_outcomes_run_id
    ON compute.predicted_game_outcomes USING btree (run_id);

CREATE INDEX idx_derived_predicted_game_outcomes_team1_id
    ON compute.predicted_game_outcomes USING btree (team1_id);

CREATE INDEX idx_derived_predicted_game_outcomes_team2_id
    ON compute.predicted_game_outcomes USING btree (team2_id);

CREATE INDEX idx_silver_predicted_game_outcomes_tournament_id
    ON compute.predicted_game_outcomes USING btree (tournament_id);

CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_legacy_matchup
    ON compute.predicted_game_outcomes USING btree (tournament_id, game_id, team1_id, team2_id)
    WHERE ((run_id IS NULL) AND (deleted_at IS NULL));

CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_run_matchup
    ON compute.predicted_game_outcomes USING btree (run_id, game_id, team1_id, team2_id)
    WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));

CREATE TRIGGER trg_derived_predicted_game_outcomes_updated_at
    BEFORE UPDATE ON compute.predicted_game_outcomes
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE ONLY compute.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_run_id_fkey
    FOREIGN KEY (run_id) REFERENCES compute.game_outcome_runs(id);

ALTER TABLE ONLY compute.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey
    FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE ONLY compute.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey
    FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE ONLY compute.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
