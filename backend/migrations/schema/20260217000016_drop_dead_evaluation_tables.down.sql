-- Best-effort recreation of dropped objects

CREATE TABLE IF NOT EXISTS derived.prediction_models (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    kind text NOT NULL,
    name text NOT NULL,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);
CREATE INDEX IF NOT EXISTS prediction_models_kind_idx ON derived.prediction_models USING btree (kind) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX IF NOT EXISTS prediction_models_kind_name_uniq ON derived.prediction_models USING btree (kind, name) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS derived.calcutta_evaluation_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    simulated_tournament_id uuid NOT NULL,
    calcutta_snapshot_id uuid,
    purpose text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid NOT NULL DEFAULT public.uuid_generate_v4(),
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    git_sha text,
    simulated_calcutta_id uuid
);

CREATE TABLE IF NOT EXISTS derived.entry_performance (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    run_id character varying(255) NOT NULL,
    entry_name character varying(255) NOT NULL,
    mean_normalized_payout double precision NOT NULL,
    median_normalized_payout double precision NOT NULL,
    p_top1 double precision NOT NULL,
    p_in_money double precision NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    calcutta_evaluation_run_id uuid
);

ALTER TABLE IF EXISTS derived.game_outcome_runs
    ADD COLUMN IF NOT EXISTS prediction_model_id uuid;
ALTER TABLE IF EXISTS derived.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_prediction_model_id_fkey
    FOREIGN KEY (prediction_model_id) REFERENCES derived.prediction_models(id);

CREATE OR REPLACE VIEW lab.entry_evaluations AS
SELECT e.id AS entry_id,
    im.name AS model_name,
    im.kind AS model_kind,
    c.name AS calcutta_name,
    e.starting_state_key,
    e.game_outcome_kind,
    e.optimizer_kind,
    ev.n_sims,
    ev.seed,
    ev.mean_normalized_payout,
    ev.median_normalized_payout,
    ev.p_top1,
    ev.p_in_money,
    ev.our_rank,
    ev.created_at AS eval_created_at
FROM (((lab.entries e
    JOIN lab.investment_models im ON ((im.id = e.investment_model_id)))
    JOIN core.calcuttas c ON ((c.id = e.calcutta_id)))
    LEFT JOIN lab.evaluations ev ON (((ev.entry_id = e.id) AND (ev.deleted_at IS NULL))))
WHERE ((e.deleted_at IS NULL) AND (im.deleted_at IS NULL));
