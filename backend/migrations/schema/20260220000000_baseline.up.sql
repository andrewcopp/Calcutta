-- =============================================================================
-- Calcutta Platform: Consolidated Baseline Schema
-- =============================================================================
-- This is the authoritative schema definition. All prior incremental migrations
-- have been folded into this single file.
-- =============================================================================

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- =============================================================================
-- SCHEMAS
-- =============================================================================

CREATE SCHEMA core;
CREATE SCHEMA derived;
CREATE SCHEMA lab;

-- =============================================================================
-- EXTENSIONS
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';

-- =============================================================================
-- FUNCTIONS: core
-- =============================================================================

CREATE FUNCTION core.calcutta_points_for_progress(p_calcutta_id uuid, p_wins integer, p_byes integer DEFAULT 0) RETURNS integer
    LANGUAGE sql STABLE
    AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.calcutta_scoring_rules r
    WHERE r.calcutta_id = p_calcutta_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;

CREATE FUNCTION core.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION core.immutable_created_by() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF OLD.created_by IS DISTINCT FROM NEW.created_by THEN
        RAISE EXCEPTION 'created_by is immutable and cannot be changed';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION core.prevent_calcutta_soft_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'Calcuttas cannot be deleted';
    END IF;
    RETURN NEW;
END;
$$;

-- =============================================================================
-- TABLES: core
-- =============================================================================

SET default_tablespace = '';
SET default_table_access_method = heap;

-- core.users
CREATE TABLE core.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255),
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    password_hash text,
    status text DEFAULT 'active'::text NOT NULL,
    invite_token_hash text,
    invite_expires_at timestamp with time zone,
    invite_consumed_at timestamp with time zone,
    invited_at timestamp with time zone,
    last_invite_sent_at timestamp with time zone,
    external_provider character varying(50),
    external_provider_id character varying(255),
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'invited'::text, 'requires_password_setup'::text, 'stub'::text])))
);

-- core.api_keys
CREATE TABLE core.api_keys (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    key_hash text NOT NULL,
    label text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone,
    last_used_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.auth_sessions
CREATE TABLE core.auth_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    refresh_token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone,
    last_used_at timestamp with time zone,
    user_agent text,
    ip_address text,
    deleted_at timestamp with time zone
);

-- core.bundle_uploads
CREATE TABLE core.bundle_uploads (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    filename text NOT NULL,
    sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    archive bytea NOT NULL,
    import_report jsonb,
    verify_report jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    status text DEFAULT 'pending'::text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    error_message text,
    CONSTRAINT bundle_uploads_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

-- core.competitions
CREATE TABLE core.competitions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.seasons
CREATE TABLE core.seasons (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    year integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.tournaments (no name column -- derived from competition + season)
CREATE TABLE core.tournaments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    competition_id uuid NOT NULL,
    season_id uuid NOT NULL,
    import_key text NOT NULL,
    rounds integer NOT NULL,
    starting_at timestamp with time zone,
    final_four_top_left character varying(50),
    final_four_bottom_left character varying(50),
    final_four_top_right character varying(50),
    final_four_bottom_right character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.schools
CREATE TABLE core.schools (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.teams
CREATE TABLE core.teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    school_id uuid NOT NULL,
    seed integer NOT NULL,
    region character varying(50) NOT NULL,
    byes integer DEFAULT 0 NOT NULL,
    wins integer DEFAULT 0 NOT NULL,
    eliminated boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_teams_seed_range CHECK (((seed >= 1) AND (seed <= 16))),
    CONSTRAINT chk_teams_byes_range CHECK (((byes >= 0) AND (byes <= 1))),
    CONSTRAINT chk_teams_wins_range CHECK (((wins >= 0) AND (wins <= 7)))
);

-- core.team_kenpom_stats
CREATE TABLE core.team_kenpom_stats (
    team_id uuid NOT NULL,
    net_rtg double precision,
    o_rtg double precision,
    d_rtg double precision,
    adj_t double precision,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.calcuttas (no bidding_open, bidding_locked_at; max_bid renamed to max_bid_points)
CREATE TABLE core.calcuttas (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    created_by uuid NOT NULL,
    visibility text DEFAULT 'private'::text NOT NULL,
    min_teams integer DEFAULT 3 NOT NULL,
    max_teams integer DEFAULT 10 NOT NULL,
    max_bid_points integer DEFAULT 50 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    budget_points integer DEFAULT 100 NOT NULL,
    CONSTRAINT chk_calcuttas_budget_positive CHECK ((budget_points > 0)),
    CONSTRAINT chk_calcuttas_max_bid_points_positive CHECK ((max_bid_points > 0)),
    CONSTRAINT chk_calcuttas_max_teams_gte_min CHECK ((max_teams >= min_teams)),
    CONSTRAINT chk_calcuttas_min_teams CHECK ((min_teams >= 1)),
    CONSTRAINT ck_core_calcuttas_max_bid_points_le_budget CHECK ((max_bid_points <= budget_points)),
    CONSTRAINT ck_calcuttas_visibility CHECK ((visibility = ANY (ARRAY['public'::text, 'unlisted'::text, 'private'::text])))
);

-- core.calcutta_scoring_rules
CREATE TABLE core.calcutta_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_scoring_rules_points_nonneg CHECK ((points_awarded >= 0)),
    CONSTRAINT chk_scoring_rules_win_index_nonneg CHECK ((win_index >= 0))
);

-- core.calcutta_invitations (revoked_at, updated status constraint)
CREATE TABLE core.calcutta_invitations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    user_id uuid NOT NULL,
    invited_by uuid NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_calcutta_invitations_status CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'revoked'::text])))
);

-- core.entries (no status column)
CREATE TABLE core.entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    user_id uuid,
    calcutta_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.entry_teams
CREATE TABLE core.entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_entry_teams_bid_positive CHECK ((bid_points > 0))
);

-- core.payouts
CREATE TABLE core.payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_payouts_amount_nonneg CHECK ((amount_cents >= 0)),
    CONSTRAINT chk_payouts_position_positive CHECK (("position" >= 1))
);

-- core.permissions
CREATE TABLE core.permissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.labels
CREATE TABLE core.labels (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.label_permissions
CREATE TABLE core.label_permissions (
    label_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- core.grants (no granted_by column)
CREATE TABLE core.grants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    scope_type text NOT NULL,
    scope_id uuid,
    label_id uuid,
    permission_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT grants_scope_id_check CHECK ((((scope_type = 'global'::text) AND (scope_id IS NULL)) OR ((scope_type <> 'global'::text) AND (scope_id IS NOT NULL)))),
    CONSTRAINT grants_scope_type_check CHECK ((scope_type = ANY (ARRAY['global'::text, 'calcutta'::text, 'tournament'::text]))),
    CONSTRAINT grants_subject_check CHECK ((((label_id IS NOT NULL) AND (permission_id IS NULL)) OR ((label_id IS NULL) AND (permission_id IS NOT NULL))))
);

-- core.idempotency_keys
CREATE TABLE core.idempotency_keys (
    key text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    response_status integer,
    response_body jsonb
);

-- =============================================================================
-- TABLES: derived
-- =============================================================================

-- derived.game_outcome_runs
CREATE TABLE derived.game_outcome_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    git_sha text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid DEFAULT public.uuid_generate_v4() NOT NULL
);

-- derived.predicted_game_outcomes (run_id is NOT NULL)
CREATE TABLE derived.predicted_game_outcomes (
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
    run_id uuid NOT NULL
);

-- derived.prediction_batches
CREATE TABLE derived.prediction_batches (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    probability_source_key text DEFAULT 'kenpom'::text NOT NULL,
    game_outcome_spec_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

COMMENT ON TABLE derived.prediction_batches IS 'Stores metadata for prediction generation runs (analogous to simulated_tournaments for simulations)';
COMMENT ON COLUMN derived.prediction_batches.probability_source_key IS 'Identifier for the probability model used (e.g., kenpom)';
COMMENT ON COLUMN derived.prediction_batches.game_outcome_spec_json IS 'Parameters for the game outcome model (e.g., {"kind": "kenpom", "sigma": 10.0})';

-- derived.predicted_team_values
CREATE TABLE derived.predicted_team_values (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    prediction_batch_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    team_id uuid NOT NULL,
    expected_points double precision NOT NULL,
    variance_points double precision,
    std_points double precision,
    p_round_1 double precision,
    p_round_2 double precision,
    p_round_3 double precision,
    p_round_4 double precision,
    p_round_5 double precision,
    p_round_6 double precision,
    p_round_7 double precision,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

COMMENT ON TABLE derived.predicted_team_values IS 'Stores predicted expected points and round probabilities for each team (analogous to simulated_teams for simulations)';
COMMENT ON COLUMN derived.predicted_team_values.expected_points IS 'Expected tournament points for 100% ownership of this team';
COMMENT ON COLUMN derived.predicted_team_values.p_round_1 IS 'Probability of winning first game (First Four or Round of 64)';
COMMENT ON COLUMN derived.predicted_team_values.p_round_2 IS 'Probability of reaching Round of 32';
COMMENT ON COLUMN derived.predicted_team_values.p_round_3 IS 'Probability of reaching Sweet 16';
COMMENT ON COLUMN derived.predicted_team_values.p_round_4 IS 'Probability of reaching Elite 8';
COMMENT ON COLUMN derived.predicted_team_values.p_round_5 IS 'Probability of reaching Final Four';
COMMENT ON COLUMN derived.predicted_team_values.p_round_6 IS 'Probability of reaching Championship game';
COMMENT ON COLUMN derived.predicted_team_values.p_round_7 IS 'Probability of winning Championship';

-- derived.run_jobs
CREATE TABLE derived.run_jobs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid NOT NULL,
    status text DEFAULT 'queued'::text NOT NULL,
    attempt integer DEFAULT 0 NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    progress_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    progress_updated_at timestamp with time zone,
    claimed_at timestamp with time zone,
    claimed_by text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_derived_run_jobs_status CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

-- derived.simulation_states
CREATE TABLE derived.simulation_states (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    source text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- derived.simulation_state_teams
CREATE TABLE derived.simulation_state_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulation_state_id uuid NOT NULL,
    team_id uuid NOT NULL,
    wins integer NOT NULL,
    byes integer NOT NULL,
    eliminated boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- derived.simulated_tournaments
CREATE TABLE derived.simulated_tournaments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    simulation_state_id uuid NOT NULL,
    n_sims integer NOT NULL,
    seed integer NOT NULL,
    probability_source_key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_simulated_tournaments_seed_nonzero CHECK ((seed <> 0))
);

-- derived.simulated_teams (simulated_tournament_id is NOT NULL)
CREATE TABLE derived.simulated_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    sim_id integer NOT NULL,
    team_id uuid NOT NULL,
    wins integer NOT NULL,
    byes integer DEFAULT 0 NOT NULL,
    eliminated boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    simulated_tournament_id uuid NOT NULL
);

-- =============================================================================
-- VIEWS: derived
-- =============================================================================

CREATE VIEW derived.portfolio_teams AS
 WITH entry_bids AS (
         SELECT ce.id AS entry_id,
            ce.calcutta_id,
            cet.team_id,
            (cet.bid_points)::double precision AS bid_points,
            cet.created_at AS entry_team_created_at,
            cet.updated_at AS entry_team_updated_at,
            sum((cet.bid_points)::double precision) OVER (PARTITION BY ce.calcutta_id, cet.team_id) AS team_total_bid_points,
            tt.school_id,
            tt.tournament_id,
            tt.seed,
            tt.region,
            tt.byes,
            tt.wins,
            tt.eliminated,
            tt.created_at AS team_created_at,
            tt.updated_at AS team_updated_at,
            t.rounds AS tournament_rounds,
            s.name AS school_name,
            GREATEST(ce.updated_at, cet.updated_at, tt.updated_at) AS derived_updated_at
           FROM ((((core.entries ce
             JOIN core.entry_teams cet ON (((cet.entry_id = ce.id) AND (cet.deleted_at IS NULL))))
             JOIN core.teams tt ON (((tt.id = cet.team_id) AND (tt.deleted_at IS NULL))))
             JOIN core.tournaments t ON (((t.id = tt.tournament_id) AND (t.deleted_at IS NULL))))
             LEFT JOIN core.schools s ON (((s.id = tt.school_id) AND (s.deleted_at IS NULL))))
          WHERE (ce.deleted_at IS NULL)
        ), entry_team_points AS (
         SELECT eb.entry_id,
            eb.calcutta_id,
            eb.team_id,
            eb.bid_points,
            eb.team_total_bid_points,
                CASE
                    WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
                    ELSE (0)::double precision
                END AS ownership_percentage,
            (
                CASE
                    WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
                    ELSE (0)::double precision
                END * (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision) AS actual_points,
            (
                CASE
                    WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
                    ELSE (0)::double precision
                END *
                CASE
                    WHEN (eb.eliminated = true) THEN (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision
                    ELSE (core.calcutta_points_for_progress(eb.calcutta_id, eb.tournament_rounds, 0))::double precision
                END) AS expected_points,
            eb.school_id,
            eb.tournament_id,
            eb.seed,
            eb.region,
            eb.byes,
            eb.wins,
            eb.eliminated,
            eb.team_created_at,
            eb.team_updated_at,
            eb.school_name,
            eb.entry_team_created_at AS created_at,
            eb.derived_updated_at AS updated_at,
            NULL::timestamp with time zone AS deleted_at,
            eb.entry_id AS portfolio_id
           FROM entry_bids eb
        )
 SELECT concat(entry_id, '-', team_id) AS id,
    entry_id AS portfolio_id,
    team_id,
    ownership_percentage,
    actual_points,
    expected_points,
    created_at,
    updated_at,
    deleted_at
   FROM entry_team_points etp;

CREATE VIEW derived.portfolios AS
 WITH entry_totals AS (
         SELECT dpt.portfolio_id AS entry_id,
            sum(dpt.expected_points) AS maximum_points,
            max(dpt.updated_at) AS updated_at
           FROM derived.portfolio_teams dpt
          GROUP BY dpt.portfolio_id
        )
 SELECT ce.id,
    ce.id AS entry_id,
    COALESCE(et.maximum_points, (0)::double precision) AS maximum_points,
    ce.created_at,
    COALESCE(et.updated_at, ce.updated_at) AS updated_at,
    NULL::timestamp with time zone AS deleted_at
   FROM (core.entries ce
     LEFT JOIN entry_totals et ON ((et.entry_id = ce.id)))
  WHERE (ce.deleted_at IS NULL);

-- =============================================================================
-- TABLES: lab
-- =============================================================================

-- lab.investment_models
CREATE TABLE lab.investment_models (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    kind text NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- lab.entries
CREATE TABLE lab.entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    investment_model_id uuid NOT NULL,
    calcutta_id uuid NOT NULL,
    game_outcome_kind text DEFAULT 'kenpom'::text NOT NULL,
    game_outcome_params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    optimizer_kind text DEFAULT 'minlp'::text NOT NULL,
    optimizer_params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    starting_state_key text DEFAULT 'post_first_four'::text NOT NULL,
    bids_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    predictions_json jsonb,
    CONSTRAINT ck_lab_entries_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['pre_tournament'::text, 'post_first_four'::text, 'current'::text])))
);

COMMENT ON COLUMN lab.entries.bids_json IS 'Optimized bids: [{team_id, bid_points, expected_roi}]. Our optimal allocation given predictions.';
COMMENT ON COLUMN lab.entries.predictions_json IS 'Market predictions: [{team_id, predicted_market_share, expected_points}]. What model predicts others will bid.';

-- lab.evaluations
CREATE TABLE lab.evaluations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    entry_id uuid NOT NULL,
    n_sims integer NOT NULL,
    seed integer NOT NULL,
    mean_normalized_payout double precision,
    median_normalized_payout double precision,
    p_top1 double precision,
    p_in_money double precision,
    our_rank integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_lab_evaluations_n_sims CHECK ((n_sims > 0)),
    CONSTRAINT ck_lab_evaluations_seed CHECK ((seed <> 0))
);

-- lab.evaluation_entry_results
CREATE TABLE lab.evaluation_entry_results (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    evaluation_id uuid NOT NULL,
    entry_name text NOT NULL,
    mean_normalized_payout double precision,
    p_top1 double precision,
    p_in_money double precision,
    rank integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

-- lab.pipeline_runs (status includes 'partial')
CREATE TABLE lab.pipeline_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    investment_model_id uuid NOT NULL,
    target_calcutta_ids uuid[] NOT NULL,
    budget_points integer DEFAULT 10000 NOT NULL,
    optimizer_kind text DEFAULT 'predicted_market_share'::text NOT NULL,
    n_sims integer DEFAULT 10000 NOT NULL,
    seed integer DEFAULT 42 NOT NULL,
    excluded_entry_name text,
    status text DEFAULT 'pending'::text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_lab_pipeline_runs_budget_points CHECK ((budget_points > 0)),
    CONSTRAINT ck_lab_pipeline_runs_n_sims CHECK ((n_sims > 0)),
    CONSTRAINT ck_lab_pipeline_runs_status CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'cancelled'::text, 'partial'::text])))
);

-- lab.pipeline_calcutta_runs
CREATE TABLE lab.pipeline_calcutta_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    pipeline_run_id uuid NOT NULL,
    calcutta_id uuid NOT NULL,
    entry_id uuid,
    stage text DEFAULT 'predictions'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    progress double precision DEFAULT 0.0 NOT NULL,
    progress_message text,
    predictions_job_id uuid,
    optimization_job_id uuid,
    evaluation_job_id uuid,
    evaluation_id uuid,
    error_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_lab_pipeline_calcutta_runs_progress CHECK (((progress >= (0.0)::double precision) AND (progress <= (1.0)::double precision))),
    CONSTRAINT ck_lab_pipeline_calcutta_runs_stage CHECK ((stage = ANY (ARRAY['predictions'::text, 'optimization'::text, 'evaluation'::text, 'completed'::text]))),
    CONSTRAINT ck_lab_pipeline_calcutta_runs_status CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

-- =============================================================================
-- VIEW: lab.model_leaderboard
-- =============================================================================

CREATE VIEW lab.model_leaderboard AS
 SELECT im.id AS investment_model_id,
    im.name AS model_name,
    im.kind AS model_kind,
    count(DISTINCT e.id) AS n_entries,
    count(DISTINCT
        CASE
            WHEN (e.predictions_json IS NOT NULL) THEN e.id
            ELSE NULL::uuid
        END) AS n_entries_with_predictions,
    count(ev.id) AS n_evaluations,
    count(DISTINCT e.calcutta_id) AS n_calcuttas_with_entries,
    count(DISTINCT
        CASE
            WHEN (ev.id IS NOT NULL) THEN e.calcutta_id
            ELSE NULL::uuid
        END) AS n_calcuttas_with_evaluations,
    avg(ev.mean_normalized_payout) AS avg_mean_payout,
    avg(ev.median_normalized_payout) AS avg_median_payout,
    avg(ev.p_top1) AS avg_p_top1,
    avg(ev.p_in_money) AS avg_p_in_money,
    min(ev.created_at) AS first_eval_at,
    max(ev.created_at) AS last_eval_at
   FROM ((lab.investment_models im
     LEFT JOIN lab.entries e ON (((e.investment_model_id = im.id) AND (e.deleted_at IS NULL))))
     LEFT JOIN lab.evaluations ev ON (((ev.entry_id = e.id) AND (ev.deleted_at IS NULL))))
  WHERE (im.deleted_at IS NULL)
  GROUP BY im.id, im.name, im.kind
  ORDER BY (avg(ev.mean_normalized_payout)) DESC NULLS LAST;

-- =============================================================================
-- PRIMARY KEYS: core
-- =============================================================================

ALTER TABLE ONLY core.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.api_keys ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.auth_sessions ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.bundle_uploads ADD CONSTRAINT bundle_uploads_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.competitions ADD CONSTRAINT competitions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.seasons ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.schools ADD CONSTRAINT schools_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.team_kenpom_stats ADD CONSTRAINT team_kenpom_stats_pkey PRIMARY KEY (team_id);
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.payouts ADD CONSTRAINT payouts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.permissions ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.labels ADD CONSTRAINT labels_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_pkey PRIMARY KEY (label_id, permission_id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.idempotency_keys ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (key, user_id);

-- =============================================================================
-- PRIMARY KEYS: derived
-- =============================================================================

ALTER TABLE ONLY derived.game_outcome_runs ADD CONSTRAINT game_outcome_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.prediction_batches ADD CONSTRAINT prediction_batches_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.predicted_team_values ADD CONSTRAINT predicted_team_values_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.run_jobs ADD CONSTRAINT run_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_states ADD CONSTRAINT simulation_states_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT simulation_state_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT simulated_tournaments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_pkey PRIMARY KEY (id);

-- =============================================================================
-- PRIMARY KEYS: lab
-- =============================================================================

ALTER TABLE ONLY lab.investment_models ADD CONSTRAINT investment_models_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.entries ADD CONSTRAINT lab_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.evaluations ADD CONSTRAINT evaluations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.evaluation_entry_results ADD CONSTRAINT evaluation_entry_results_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.pipeline_runs ADD CONSTRAINT pipeline_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_pkey PRIMARY KEY (id);

-- =============================================================================
-- UNIQUE CONSTRAINTS: core
-- =============================================================================

ALTER TABLE ONLY core.permissions ADD CONSTRAINT permissions_key_key UNIQUE (key);
ALTER TABLE ONLY core.labels ADD CONSTRAINT labels_key_key UNIQUE (key);
ALTER TABLE ONLY core.competitions ADD CONSTRAINT uq_core_competitions_name UNIQUE (name);
ALTER TABLE ONLY core.seasons ADD CONSTRAINT uq_core_seasons_year UNIQUE (year);
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT uq_core_calcutta_scoring_rules UNIQUE (calcutta_id, win_index);
ALTER TABLE ONLY core.payouts ADD CONSTRAINT uq_core_payouts_calcutta_position UNIQUE (calcutta_id, "position");

-- =============================================================================
-- UNIQUE CONSTRAINTS: derived
-- =============================================================================

ALTER TABLE ONLY derived.simulation_state_teams
    ADD CONSTRAINT uq_derived_simulation_state_teams_state_team UNIQUE (simulation_state_id, team_id);

ALTER TABLE ONLY lab.evaluation_entry_results
    ADD CONSTRAINT evaluation_entry_results_evaluation_id_entry_name_key UNIQUE (evaluation_id, entry_name);

-- =============================================================================
-- INDEXES: core.users
-- =============================================================================

CREATE UNIQUE INDEX uq_users_email ON core.users USING btree (email) WHERE ((email IS NOT NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_users_external_provider ON core.users USING btree (external_provider, external_provider_id) WHERE ((external_provider_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_users_invite_token_hash ON core.users USING btree (invite_token_hash) WHERE (invite_token_hash IS NOT NULL);
CREATE UNIQUE INDEX users_stub_name_unique ON core.users USING btree (first_name, last_name) WHERE ((status = 'stub'::text) AND (deleted_at IS NULL));
CREATE INDEX idx_users_invite_expires_at ON core.users USING btree (invite_expires_at);
CREATE INDEX idx_users_status ON core.users USING btree (status);

-- =============================================================================
-- INDEXES: core.api_keys
-- =============================================================================

CREATE UNIQUE INDEX uq_api_keys_key_hash ON core.api_keys USING btree (key_hash);
CREATE INDEX idx_api_keys_user_id ON core.api_keys USING btree (user_id);
CREATE INDEX idx_api_keys_revoked_at ON core.api_keys USING btree (revoked_at);

-- =============================================================================
-- INDEXES: core.auth_sessions
-- =============================================================================

CREATE UNIQUE INDEX uq_auth_sessions_refresh_token_hash ON core.auth_sessions USING btree (refresh_token_hash);
CREATE INDEX idx_auth_sessions_user_id ON core.auth_sessions USING btree (user_id);
CREATE INDEX idx_auth_sessions_expires_at ON core.auth_sessions USING btree (expires_at);
CREATE INDEX idx_auth_sessions_revoked_at ON core.auth_sessions USING btree (revoked_at);

-- =============================================================================
-- INDEXES: core.bundle_uploads
-- =============================================================================

CREATE UNIQUE INDEX uq_bundle_uploads_sha256 ON core.bundle_uploads USING btree (sha256) WHERE (deleted_at IS NULL);
CREATE INDEX idx_bundle_uploads_created_at ON core.bundle_uploads USING btree (created_at);
CREATE INDEX idx_bundle_uploads_status_created_at ON core.bundle_uploads USING btree (status, created_at) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: core.tournaments
-- =============================================================================

CREATE UNIQUE INDEX uq_core_tournaments_import_key ON core.tournaments USING btree (import_key) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX idx_tournaments_competition_season_active ON core.tournaments USING btree (competition_id, season_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_tournaments_competition_id ON core.tournaments USING btree (competition_id);
CREATE INDEX idx_core_tournaments_season_id ON core.tournaments USING btree (season_id);

-- =============================================================================
-- INDEXES: core.schools
-- =============================================================================

CREATE UNIQUE INDEX uq_core_schools_name ON core.schools USING btree (name) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_core_schools_slug ON core.schools USING btree (slug) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: core.teams
-- =============================================================================

CREATE UNIQUE INDEX idx_teams_tournament_school_active ON core.teams USING btree (tournament_id, school_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_teams_tournament_id ON core.teams USING btree (tournament_id);
CREATE INDEX idx_core_teams_school_id ON core.teams USING btree (school_id);
CREATE INDEX idx_teams_tournament_region_seed_active ON core.teams USING btree (tournament_id, region, seed) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: core.calcuttas
-- =============================================================================

CREATE INDEX idx_core_calcuttas_tournament_id ON core.calcuttas USING btree (tournament_id);
CREATE INDEX idx_core_calcuttas_owner_id ON core.calcuttas USING btree (owner_id);
CREATE INDEX idx_calcuttas_created_by ON core.calcuttas USING btree (created_by);

-- =============================================================================
-- INDEXES: core.calcutta_scoring_rules
-- =============================================================================

CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id ON core.calcutta_scoring_rules USING btree (calcutta_id);

-- =============================================================================
-- INDEXES: core.calcutta_invitations
-- =============================================================================

CREATE UNIQUE INDEX uq_calcutta_invitations_calcutta_user ON core.calcutta_invitations USING btree (calcutta_id, user_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcutta_invitations_calcutta_id_active ON core.calcutta_invitations USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcutta_invitations_user_id ON core.calcutta_invitations USING btree (user_id);
CREATE INDEX idx_calcutta_invitations_invited_by ON core.calcutta_invitations USING btree (invited_by);

-- =============================================================================
-- INDEXES: core.entries
-- =============================================================================

CREATE UNIQUE INDEX uq_entries_user_calcutta ON core.entries USING btree (user_id, calcutta_id) WHERE ((user_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_entries_calcutta_id_active ON core.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_entries_user_id ON core.entries USING btree (user_id);

-- =============================================================================
-- INDEXES: core.entry_teams
-- =============================================================================

CREATE UNIQUE INDEX uq_core_entry_teams_entry_team ON core.entry_teams USING btree (entry_id, team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_entry_teams_entry_id_active ON core.entry_teams USING btree (entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_entry_teams_team_id_active ON core.entry_teams USING btree (team_id) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: core.payouts
-- =============================================================================

CREATE INDEX idx_core_payouts_calcutta_id ON core.payouts USING btree (calcutta_id);

-- =============================================================================
-- INDEXES: core.label_permissions
-- =============================================================================

CREATE INDEX idx_label_permissions_label_id ON core.label_permissions USING btree (label_id);
CREATE INDEX idx_label_permissions_permission_id ON core.label_permissions USING btree (permission_id);

-- =============================================================================
-- INDEXES: core.grants
-- =============================================================================

CREATE UNIQUE INDEX uq_grants_user_label_scope_active ON core.grants USING btree (user_id, label_id, scope_type, scope_id) WHERE ((label_id IS NOT NULL) AND (revoked_at IS NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_grants_user_permission_scope_active ON core.grants USING btree (user_id, permission_id, scope_type, scope_id) WHERE ((permission_id IS NOT NULL) AND (revoked_at IS NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_grants_user_scope_active ON core.grants USING btree (user_id, scope_type, scope_id) WHERE ((revoked_at IS NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_grants_user_id ON core.grants USING btree (user_id);
CREATE INDEX idx_grants_label_id ON core.grants USING btree (label_id);
CREATE INDEX idx_grants_permission_id ON core.grants USING btree (permission_id);
CREATE INDEX idx_grants_scope ON core.grants USING btree (scope_type, scope_id);
CREATE INDEX idx_grants_revoked_at ON core.grants USING btree (revoked_at);

-- =============================================================================
-- INDEXES: core.idempotency_keys
-- =============================================================================

CREATE INDEX idx_idempotency_keys_created ON core.idempotency_keys USING btree (created_at);

-- =============================================================================
-- INDEXES: derived.game_outcome_runs
-- =============================================================================

CREATE INDEX idx_derived_game_outcome_runs_tournament_id ON derived.game_outcome_runs USING btree (tournament_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_game_outcome_runs_run_key ON derived.game_outcome_runs USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_game_outcome_runs_created_at ON derived.game_outcome_runs USING btree (created_at DESC) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: derived.predicted_game_outcomes
-- =============================================================================

CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_run_matchup ON derived.predicted_game_outcomes USING btree (run_id, game_id, team1_id, team2_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_predicted_game_outcomes_run_id ON derived.predicted_game_outcomes USING btree (run_id);
CREATE INDEX idx_derived_predicted_game_outcomes_team1_id ON derived.predicted_game_outcomes USING btree (team1_id);
CREATE INDEX idx_derived_predicted_game_outcomes_team2_id ON derived.predicted_game_outcomes USING btree (team2_id);
CREATE INDEX idx_silver_predicted_game_outcomes_tournament_id ON derived.predicted_game_outcomes USING btree (tournament_id);

-- =============================================================================
-- INDEXES: derived.prediction_batches
-- =============================================================================

CREATE INDEX idx_prediction_batches_tournament_id ON derived.prediction_batches USING btree (tournament_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_prediction_batches_created_at ON derived.prediction_batches USING btree (created_at DESC) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: derived.predicted_team_values
-- =============================================================================

CREATE INDEX idx_predicted_team_values_batch_id ON derived.predicted_team_values USING btree (prediction_batch_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_predicted_team_values_team_id ON derived.predicted_team_values USING btree (team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_predicted_team_values_tournament_id ON derived.predicted_team_values USING btree (tournament_id) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: derived.run_jobs
-- =============================================================================

CREATE UNIQUE INDEX uq_derived_run_jobs_kind_run_id ON derived.run_jobs USING btree (run_kind, run_id);
CREATE INDEX idx_derived_run_jobs_kind_status_created_at ON derived.run_jobs USING btree (run_kind, status, created_at);
CREATE INDEX idx_derived_run_jobs_progress_updated_at ON derived.run_jobs USING btree (progress_updated_at);

-- =============================================================================
-- INDEXES: derived.simulation_states
-- =============================================================================

CREATE INDEX idx_analytics_tournament_state_snapshots_tournament_id ON derived.simulation_states USING btree (tournament_id);

-- =============================================================================
-- INDEXES: derived.simulation_state_teams
-- =============================================================================

CREATE INDEX idx_analytics_tournament_state_snapshot_teams_snapshot_id ON derived.simulation_state_teams USING btree (simulation_state_id);
CREATE INDEX idx_analytics_tournament_state_snapshot_teams_team_id ON derived.simulation_state_teams USING btree (team_id);

-- =============================================================================
-- INDEXES: derived.simulated_tournaments
-- =============================================================================

CREATE UNIQUE INDEX uq_analytics_tournament_simulation_batches_natural_key ON derived.simulated_tournaments USING btree (tournament_id, simulation_state_id, n_sims, seed, probability_source_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_analytics_tournament_simulation_batches_snapshot_id ON derived.simulated_tournaments USING btree (simulation_state_id);
CREATE INDEX idx_analytics_tournament_simulation_batches_tournament_id ON derived.simulated_tournaments USING btree (tournament_id);

-- =============================================================================
-- INDEXES: derived.simulated_teams
-- =============================================================================

CREATE UNIQUE INDEX uq_analytics_simulated_tournaments_batch_sim_team ON derived.simulated_teams USING btree (simulated_tournament_id, sim_id, team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_analytics_simulated_tournaments_batch_id ON derived.simulated_teams USING btree (simulated_tournament_id);
CREATE INDEX idx_silver_simulated_tournaments_sim_id ON derived.simulated_teams USING btree (tournament_id, sim_id);
CREATE INDEX idx_silver_simulated_tournaments_team_id ON derived.simulated_teams USING btree (team_id);
CREATE INDEX idx_silver_simulated_tournaments_tournament_id ON derived.simulated_teams USING btree (tournament_id);

-- =============================================================================
-- INDEXES: lab.investment_models
-- =============================================================================

CREATE UNIQUE INDEX uq_lab_investment_models_name ON lab.investment_models USING btree (name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_investment_models_kind ON lab.investment_models USING btree (kind) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_investment_models_created_at ON lab.investment_models USING btree (created_at DESC) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: lab.entries
-- =============================================================================

CREATE UNIQUE INDEX uq_lab_entries_model_calcutta_state ON lab.entries USING btree (investment_model_id, calcutta_id, starting_state_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_investment_model_id ON lab.entries USING btree (investment_model_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_calcutta_id ON lab.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_created_at ON lab.entries USING btree (created_at DESC) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: lab.evaluations
-- =============================================================================

CREATE UNIQUE INDEX uq_lab_evaluations_entry_sims_seed ON lab.evaluations USING btree (entry_id, n_sims, seed) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_evaluations_entry_id ON lab.evaluations USING btree (entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_evaluations_created_at ON lab.evaluations USING btree (created_at DESC) WHERE (deleted_at IS NULL);

-- =============================================================================
-- INDEXES: lab.evaluation_entry_results
-- =============================================================================

CREATE INDEX idx_evaluation_entry_results_evaluation_id ON lab.evaluation_entry_results USING btree (evaluation_id);

-- =============================================================================
-- INDEXES: lab.pipeline_runs
-- =============================================================================

CREATE INDEX idx_lab_pipeline_runs_investment_model_id ON lab.pipeline_runs USING btree (investment_model_id);
CREATE INDEX idx_lab_pipeline_runs_status ON lab.pipeline_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));
CREATE INDEX idx_lab_pipeline_runs_created_at ON lab.pipeline_runs USING btree (created_at DESC);

-- =============================================================================
-- INDEXES: lab.pipeline_calcutta_runs
-- =============================================================================

CREATE UNIQUE INDEX uq_lab_pipeline_calcutta_runs_pipeline_calcutta ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id, calcutta_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_pipeline_run_id ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_calcutta_id ON lab.pipeline_calcutta_runs USING btree (calcutta_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_entry_id ON lab.pipeline_calcutta_runs USING btree (entry_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_evaluation_id ON lab.pipeline_calcutta_runs USING btree (evaluation_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_status ON lab.pipeline_calcutta_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));

-- =============================================================================
-- TRIGGERS: core
-- =============================================================================

CREATE TRIGGER trg_core_users_updated_at BEFORE UPDATE ON core.users FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_api_keys_updated_at BEFORE UPDATE ON core.api_keys FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_auth_sessions_updated_at BEFORE UPDATE ON core.auth_sessions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_bundle_uploads_updated_at BEFORE UPDATE ON core.bundle_uploads FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_competitions_updated_at BEFORE UPDATE ON core.competitions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_seasons_updated_at BEFORE UPDATE ON core.seasons FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_tournaments_updated_at BEFORE UPDATE ON core.tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_schools_updated_at BEFORE UPDATE ON core.schools FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_teams_updated_at BEFORE UPDATE ON core.teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_team_kenpom_stats_updated_at BEFORE UPDATE ON core.team_kenpom_stats FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcuttas_updated_at BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_calcuttas_immutable_created_by BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.immutable_created_by();
CREATE TRIGGER trg_calcuttas_prevent_soft_delete BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.prevent_calcutta_soft_delete();
CREATE TRIGGER trg_core_calcutta_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_invitations_updated_at BEFORE UPDATE ON core.calcutta_invitations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entries_updated_at BEFORE UPDATE ON core.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entry_teams_updated_at BEFORE UPDATE ON core.entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_payouts_updated_at BEFORE UPDATE ON core.payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_permissions_updated_at BEFORE UPDATE ON core.permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_labels_updated_at BEFORE UPDATE ON core.labels FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_label_permissions_updated_at BEFORE UPDATE ON core.label_permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_grants_updated_at BEFORE UPDATE ON core.grants FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- TRIGGERS: derived
-- =============================================================================

CREATE TRIGGER trg_derived_game_outcome_runs_updated_at BEFORE UPDATE ON derived.game_outcome_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_predicted_game_outcomes_updated_at BEFORE UPDATE ON derived.predicted_game_outcomes FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_run_jobs_updated_at BEFORE UPDATE ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_states_updated_at BEFORE UPDATE ON derived.simulation_states FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_state_teams_updated_at BEFORE UPDATE ON derived.simulation_state_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_tournaments_updated_at BEFORE UPDATE ON derived.simulated_tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_teams_updated_at BEFORE UPDATE ON derived.simulated_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- TRIGGERS: lab
-- =============================================================================

CREATE TRIGGER trg_lab_investment_models_updated_at BEFORE UPDATE ON lab.investment_models FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_entries_updated_at BEFORE UPDATE ON lab.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_evaluations_updated_at BEFORE UPDATE ON lab.evaluations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_pipeline_runs_updated_at BEFORE UPDATE ON lab.pipeline_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_pipeline_calcutta_runs_updated_at BEFORE UPDATE ON lab.pipeline_calcutta_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- FOREIGN KEYS: core
-- =============================================================================

ALTER TABLE ONLY core.api_keys ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.auth_sessions ADD CONSTRAINT auth_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_competition_id_fkey FOREIGN KEY (competition_id) REFERENCES core.competitions(id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_season_id_fkey FOREIGN KEY (season_id) REFERENCES core.seasons(id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_school_id_fkey FOREIGN KEY (school_id) REFERENCES core.schools(id);
ALTER TABLE ONLY core.team_kenpom_stats ADD CONSTRAINT team_kenpom_stats_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_created_by_fkey FOREIGN KEY (created_by) REFERENCES core.users(id);
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES core.users(id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY core.payouts ADD CONSTRAINT payouts_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);

-- =============================================================================
-- FOREIGN KEYS: derived
-- =============================================================================

ALTER TABLE ONLY derived.game_outcome_runs ADD CONSTRAINT game_outcome_runs_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_run_id_fkey FOREIGN KEY (run_id) REFERENCES derived.game_outcome_runs(id);
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.prediction_batches ADD CONSTRAINT prediction_batches_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.predicted_team_values ADD CONSTRAINT predicted_team_values_prediction_batch_id_fkey FOREIGN KEY (prediction_batch_id) REFERENCES derived.prediction_batches(id);
ALTER TABLE ONLY derived.predicted_team_values ADD CONSTRAINT predicted_team_values_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.predicted_team_values ADD CONSTRAINT predicted_team_values_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY derived.simulation_states ADD CONSTRAINT simulation_states_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT simulation_state_teams_simulation_state_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT simulation_state_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT simulated_tournaments_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT simulated_tournaments_simulation_state_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id);
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE RESTRICT;
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE RESTRICT;
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_simulated_tournament_id_fkey FOREIGN KEY (simulated_tournament_id) REFERENCES derived.simulated_tournaments(id);

-- =============================================================================
-- FOREIGN KEYS: lab
-- =============================================================================

ALTER TABLE ONLY lab.entries ADD CONSTRAINT entries_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);
ALTER TABLE ONLY lab.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY lab.evaluations ADD CONSTRAINT evaluations_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);
ALTER TABLE ONLY lab.evaluation_entry_results ADD CONSTRAINT evaluation_entry_results_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);
ALTER TABLE ONLY lab.pipeline_runs ADD CONSTRAINT pipeline_runs_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_pipeline_run_id_fkey FOREIGN KEY (pipeline_run_id) REFERENCES lab.pipeline_runs(id) ON DELETE CASCADE;
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_predictions_job_id_fkey FOREIGN KEY (predictions_job_id) REFERENCES derived.run_jobs(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_optimization_job_id_fkey FOREIGN KEY (optimization_job_id) REFERENCES derived.run_jobs(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_evaluation_job_id_fkey FOREIGN KEY (evaluation_job_id) REFERENCES derived.run_jobs(id);
