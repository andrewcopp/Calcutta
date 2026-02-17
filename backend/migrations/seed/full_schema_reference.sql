--
-- PostgreSQL database dump
--

-- Dumped from database version 16.12
-- Dumped by pg_dump version 16.12

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

--
-- Name: core; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA core;


--
-- Name: derived; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA derived;


--
-- Name: lab; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA lab;


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: calcutta_points_for_progress(uuid, integer, integer); Type: FUNCTION; Schema: core; Owner: -
--

CREATE FUNCTION core.calcutta_points_for_progress(p_calcutta_id uuid, p_wins integer, p_byes integer DEFAULT 0) RETURNS integer
    LANGUAGE sql STABLE
    AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.calcutta_scoring_rules r
    WHERE r.calcutta_id = p_calcutta_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;


--
-- Name: calcutta_slugify(text); Type: FUNCTION; Schema: core; Owner: -
--

CREATE FUNCTION core.calcutta_slugify(input text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT trim(both '-' from regexp_replace(lower(input), '[^a-z0-9]+', '-', 'g'))
$$;


--
-- Name: set_updated_at(); Type: FUNCTION; Schema: core; Owner: -
--

CREATE FUNCTION core.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_keys; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: auth_sessions; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: bundle_uploads; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: calcutta_invitations; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_invitations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    user_id uuid NOT NULL,
    invited_by uuid NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_calcutta_invitations_status CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'declined'::text])))
);


--
-- Name: calcutta_scoring_rules; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: calcutta_snapshot_entries; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_snapshot_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    entry_id uuid,
    display_name text NOT NULL,
    is_synthetic boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: calcutta_snapshot_entry_teams; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_snapshot_entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: calcutta_snapshot_payouts; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_snapshot_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: calcutta_snapshot_scoring_rules; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_snapshot_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: calcutta_snapshots; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcutta_snapshots (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    base_calcutta_id uuid NOT NULL,
    snapshot_type text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: calcuttas; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.calcuttas (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    min_teams integer DEFAULT 3 NOT NULL,
    max_teams integer DEFAULT 10 NOT NULL,
    max_bid integer DEFAULT 50 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    budget_points integer DEFAULT 100 NOT NULL,
    bidding_open boolean DEFAULT true NOT NULL,
    bidding_locked_at timestamp with time zone,
    CONSTRAINT chk_calcuttas_budget_positive CHECK ((budget_points > 0)),
    CONSTRAINT chk_calcuttas_max_bid_positive CHECK ((max_bid > 0)),
    CONSTRAINT chk_calcuttas_max_teams_gte_min CHECK ((max_teams >= min_teams)),
    CONSTRAINT chk_calcuttas_min_teams CHECK ((min_teams >= 1)),
    CONSTRAINT ck_core_calcuttas_max_bid_le_budget CHECK ((max_bid <= budget_points))
);


--
-- Name: competitions; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.competitions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: entries; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    user_id uuid,
    calcutta_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: entry_teams; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: grants; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: idempotency_keys; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.idempotency_keys (
    key text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    response_status integer,
    response_body jsonb
);


--
-- Name: label_permissions; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.label_permissions (
    label_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: labels; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.labels (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: payouts; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: permissions; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.permissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: schools; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.schools (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: seasons; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.seasons (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    year integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: team_kenpom_stats; Type: TABLE; Schema: core; Owner: -
--

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


--
-- Name: teams; Type: TABLE; Schema: core; Owner: -
--

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
    CONSTRAINT chk_teams_wins_nonneg CHECK ((wins >= 0))
);


--
-- Name: tournaments; Type: TABLE; Schema: core; Owner: -
--

CREATE TABLE core.tournaments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    competition_id uuid NOT NULL,
    season_id uuid NOT NULL,
    name character varying(255) NOT NULL,
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


--
-- Name: users; Type: TABLE; Schema: core; Owner: -
--

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
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'invited'::text, 'requires_password_setup'::text, 'stub'::text])))
);


--
-- Name: calcutta_evaluation_runs; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.calcutta_evaluation_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_tournament_id uuid NOT NULL,
    calcutta_snapshot_id uuid,
    purpose text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    git_sha text,
    simulated_calcutta_id uuid
);


--
-- Name: entry_performance; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.entry_performance (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
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


--
-- Name: game_outcome_runs; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.game_outcome_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    git_sha text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    prediction_model_id uuid
);


--
-- Name: portfolio_teams; Type: VIEW; Schema: derived; Owner: -
--

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
            (0)::double precision AS predicted_points,
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
    predicted_points,
    created_at,
    updated_at,
    deleted_at
   FROM entry_team_points etp;


--
-- Name: portfolios; Type: VIEW; Schema: derived; Owner: -
--

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


--
-- Name: predicted_game_outcomes; Type: TABLE; Schema: derived; Owner: -
--

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
    run_id uuid
);


--
-- Name: prediction_models; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.prediction_models (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    kind text NOT NULL,
    name text NOT NULL,
    params_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: run_jobs; Type: TABLE; Schema: derived; Owner: -
--

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


--
-- Name: simulated_calcutta_payouts; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulated_calcutta_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: simulated_calcutta_scoring_rules; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulated_calcutta_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: simulated_calcuttas; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulated_calcuttas (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    tournament_id uuid NOT NULL,
    base_calcutta_id uuid,
    starting_state_key text DEFAULT 'post_first_four'::text NOT NULL,
    excluded_entry_name text,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_simulated_calcuttas_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text])))
);


--
-- Name: simulated_entries; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulated_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    display_name text NOT NULL,
    source_kind text NOT NULL,
    source_entry_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_simulated_entries_source_kind CHECK ((source_kind = ANY (ARRAY['manual'::text, 'from_real_entry'::text])))
);


--
-- Name: simulated_entry_teams; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulated_entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: simulated_teams; Type: TABLE; Schema: derived; Owner: -
--

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
    simulated_tournament_id uuid
);


--
-- Name: simulated_tournaments; Type: TABLE; Schema: derived; Owner: -
--

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


--
-- Name: simulation_state_teams; Type: TABLE; Schema: derived; Owner: -
--

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


--
-- Name: simulation_states; Type: TABLE; Schema: derived; Owner: -
--

CREATE TABLE derived.simulation_states (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    source text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: entries; Type: TABLE; Schema: lab; Owner: -
--

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


--
-- Name: evaluations; Type: TABLE; Schema: lab; Owner: -
--

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
    simulated_calcutta_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_lab_evaluations_n_sims CHECK ((n_sims > 0)),
    CONSTRAINT ck_lab_evaluations_seed CHECK ((seed <> 0))
);


--
-- Name: investment_models; Type: TABLE; Schema: lab; Owner: -
--

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


--
-- Name: entry_evaluations; Type: VIEW; Schema: lab; Owner: -
--

CREATE VIEW lab.entry_evaluations AS
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


--
-- Name: evaluation_entry_results; Type: TABLE; Schema: lab; Owner: -
--

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


--
-- Name: model_leaderboard; Type: VIEW; Schema: lab; Owner: -
--

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


--
-- Name: pipeline_calcutta_runs; Type: TABLE; Schema: lab; Owner: -
--

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


--
-- Name: pipeline_runs; Type: TABLE; Schema: lab; Owner: -
--

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
    CONSTRAINT ck_lab_pipeline_runs_status CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'cancelled'::text])))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: auth_sessions auth_sessions_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.auth_sessions
    ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (id);


--
-- Name: bundle_uploads bundle_uploads_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.bundle_uploads
    ADD CONSTRAINT bundle_uploads_pkey PRIMARY KEY (id);


--
-- Name: calcutta_invitations calcutta_invitations_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_invitations
    ADD CONSTRAINT calcutta_invitations_pkey PRIMARY KEY (id);


--
-- Name: calcutta_scoring_rules calcutta_scoring_rules_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_scoring_rules
    ADD CONSTRAINT calcutta_scoring_rules_pkey PRIMARY KEY (id);


--
-- Name: calcutta_snapshot_entries calcutta_snapshot_entries_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entries
    ADD CONSTRAINT calcutta_snapshot_entries_pkey PRIMARY KEY (id);


--
-- Name: calcutta_snapshot_entry_teams calcutta_snapshot_entry_teams_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entry_teams
    ADD CONSTRAINT calcutta_snapshot_entry_teams_pkey PRIMARY KEY (id);


--
-- Name: calcutta_snapshot_payouts calcutta_snapshot_payouts_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_payouts
    ADD CONSTRAINT calcutta_snapshot_payouts_pkey PRIMARY KEY (id);


--
-- Name: calcutta_snapshot_scoring_rules calcutta_snapshot_scoring_rules_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules
    ADD CONSTRAINT calcutta_snapshot_scoring_rules_pkey PRIMARY KEY (id);


--
-- Name: calcutta_snapshots calcutta_snapshots_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshots
    ADD CONSTRAINT calcutta_snapshots_pkey PRIMARY KEY (id);


--
-- Name: calcuttas calcuttas_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcuttas
    ADD CONSTRAINT calcuttas_pkey PRIMARY KEY (id);


--
-- Name: competitions competitions_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.competitions
    ADD CONSTRAINT competitions_pkey PRIMARY KEY (id);


--
-- Name: entries entries_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entries
    ADD CONSTRAINT entries_pkey PRIMARY KEY (id);


--
-- Name: entry_teams entry_teams_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entry_teams
    ADD CONSTRAINT entry_teams_pkey PRIMARY KEY (id);


--
-- Name: grants grants_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.grants
    ADD CONSTRAINT grants_pkey PRIMARY KEY (id);


--
-- Name: idempotency_keys idempotency_keys_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.idempotency_keys
    ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (key, user_id);


--
-- Name: label_permissions label_permissions_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.label_permissions
    ADD CONSTRAINT label_permissions_pkey PRIMARY KEY (label_id, permission_id);


--
-- Name: labels labels_key_key; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.labels
    ADD CONSTRAINT labels_key_key UNIQUE (key);


--
-- Name: labels labels_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.labels
    ADD CONSTRAINT labels_pkey PRIMARY KEY (id);


--
-- Name: payouts payouts_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.payouts
    ADD CONSTRAINT payouts_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_key_key; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.permissions
    ADD CONSTRAINT permissions_key_key UNIQUE (key);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: schools schools_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.schools
    ADD CONSTRAINT schools_pkey PRIMARY KEY (id);


--
-- Name: seasons seasons_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.seasons
    ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);


--
-- Name: team_kenpom_stats team_kenpom_stats_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.team_kenpom_stats
    ADD CONSTRAINT team_kenpom_stats_pkey PRIMARY KEY (team_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: tournaments tournaments_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.tournaments
    ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);


--
-- Name: calcutta_scoring_rules uq_core_calcutta_scoring_rules; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_scoring_rules
    ADD CONSTRAINT uq_core_calcutta_scoring_rules UNIQUE (calcutta_id, win_index);


--
-- Name: calcutta_snapshot_entries uq_core_calcutta_snapshot_entries_snapshot_display_name; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entries
    ADD CONSTRAINT uq_core_calcutta_snapshot_entries_snapshot_display_name UNIQUE (calcutta_snapshot_id, display_name);


--
-- Name: calcutta_snapshot_entry_teams uq_core_calcutta_snapshot_entry_teams_entry_team; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entry_teams
    ADD CONSTRAINT uq_core_calcutta_snapshot_entry_teams_entry_team UNIQUE (calcutta_snapshot_entry_id, team_id);


--
-- Name: calcutta_snapshot_payouts uq_core_calcutta_snapshot_payouts_snapshot_position; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_payouts
    ADD CONSTRAINT uq_core_calcutta_snapshot_payouts_snapshot_position UNIQUE (calcutta_snapshot_id, "position");


--
-- Name: calcutta_snapshot_scoring_rules uq_core_calcutta_snapshot_scoring_rules_snapshot_win_index; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules
    ADD CONSTRAINT uq_core_calcutta_snapshot_scoring_rules_snapshot_win_index UNIQUE (calcutta_snapshot_id, win_index);


--
-- Name: competitions uq_core_competitions_name; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.competitions
    ADD CONSTRAINT uq_core_competitions_name UNIQUE (name);


--
-- Name: payouts uq_core_payouts_calcutta_position; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.payouts
    ADD CONSTRAINT uq_core_payouts_calcutta_position UNIQUE (calcutta_id, "position");


--
-- Name: seasons uq_core_seasons_year; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.seasons
    ADD CONSTRAINT uq_core_seasons_year UNIQUE (year);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: calcutta_evaluation_runs calcutta_evaluation_runs_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.calcutta_evaluation_runs
    ADD CONSTRAINT calcutta_evaluation_runs_pkey PRIMARY KEY (id);


--
-- Name: game_outcome_runs game_outcome_runs_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_pkey PRIMARY KEY (id);


--
-- Name: entry_performance gold_entry_performance_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.entry_performance
    ADD CONSTRAINT gold_entry_performance_pkey PRIMARY KEY (id);


--
-- Name: prediction_models prediction_models_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.prediction_models
    ADD CONSTRAINT prediction_models_pkey PRIMARY KEY (id);


--
-- Name: run_jobs run_jobs_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.run_jobs
    ADD CONSTRAINT run_jobs_pkey PRIMARY KEY (id);


--
-- Name: predicted_game_outcomes silver_predicted_game_outcomes_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.predicted_game_outcomes
    ADD CONSTRAINT silver_predicted_game_outcomes_pkey PRIMARY KEY (id);


--
-- Name: simulated_teams silver_simulated_tournaments_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_teams
    ADD CONSTRAINT silver_simulated_tournaments_pkey PRIMARY KEY (id);


--
-- Name: simulated_calcutta_payouts simulated_calcutta_payouts_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcutta_payouts
    ADD CONSTRAINT simulated_calcutta_payouts_pkey PRIMARY KEY (id);


--
-- Name: simulated_calcutta_scoring_rules simulated_calcutta_scoring_rules_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcutta_scoring_rules
    ADD CONSTRAINT simulated_calcutta_scoring_rules_pkey PRIMARY KEY (id);


--
-- Name: simulated_calcuttas simulated_calcuttas_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcuttas
    ADD CONSTRAINT simulated_calcuttas_pkey PRIMARY KEY (id);


--
-- Name: simulated_entries simulated_entries_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entries
    ADD CONSTRAINT simulated_entries_pkey PRIMARY KEY (id);


--
-- Name: simulated_entry_teams simulated_entry_teams_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entry_teams
    ADD CONSTRAINT simulated_entry_teams_pkey PRIMARY KEY (id);


--
-- Name: simulated_tournaments tournament_simulation_batches_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_tournaments
    ADD CONSTRAINT tournament_simulation_batches_pkey PRIMARY KEY (id);


--
-- Name: simulation_state_teams tournament_state_snapshot_teams_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_state_teams
    ADD CONSTRAINT tournament_state_snapshot_teams_pkey PRIMARY KEY (id);


--
-- Name: simulation_states tournament_state_snapshots_pkey; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_states
    ADD CONSTRAINT tournament_state_snapshots_pkey PRIMARY KEY (id);


--
-- Name: simulation_state_teams uq_analytics_tournament_state_snapshot_teams_snapshot_team; Type: CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_state_teams
    ADD CONSTRAINT uq_analytics_tournament_state_snapshot_teams_snapshot_team UNIQUE (simulation_state_id, team_id);


--
-- Name: evaluation_entry_results evaluation_entry_results_evaluation_id_entry_name_key; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluation_entry_results
    ADD CONSTRAINT evaluation_entry_results_evaluation_id_entry_name_key UNIQUE (evaluation_id, entry_name);


--
-- Name: evaluation_entry_results evaluation_entry_results_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluation_entry_results
    ADD CONSTRAINT evaluation_entry_results_pkey PRIMARY KEY (id);


--
-- Name: evaluations evaluations_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluations
    ADD CONSTRAINT evaluations_pkey PRIMARY KEY (id);


--
-- Name: investment_models investment_models_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.investment_models
    ADD CONSTRAINT investment_models_pkey PRIMARY KEY (id);


--
-- Name: entries lab_entries_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.entries
    ADD CONSTRAINT lab_entries_pkey PRIMARY KEY (id);


--
-- Name: pipeline_calcutta_runs pipeline_calcutta_runs_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_calcutta_runs
    ADD CONSTRAINT pipeline_calcutta_runs_pkey PRIMARY KEY (id);


--
-- Name: pipeline_runs pipeline_runs_pkey; Type: CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_runs
    ADD CONSTRAINT pipeline_runs_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_api_keys_revoked_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_api_keys_revoked_at ON core.api_keys USING btree (revoked_at);


--
-- Name: idx_api_keys_user_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_api_keys_user_id ON core.api_keys USING btree (user_id);


--
-- Name: idx_auth_sessions_expires_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_auth_sessions_expires_at ON core.auth_sessions USING btree (expires_at);


--
-- Name: idx_auth_sessions_revoked_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_auth_sessions_revoked_at ON core.auth_sessions USING btree (revoked_at);


--
-- Name: idx_auth_sessions_user_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_auth_sessions_user_id ON core.auth_sessions USING btree (user_id);


--
-- Name: idx_bundle_uploads_created_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_bundle_uploads_created_at ON core.bundle_uploads USING btree (created_at);


--
-- Name: idx_bundle_uploads_status_created_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_bundle_uploads_status_created_at ON core.bundle_uploads USING btree (status, created_at) WHERE (deleted_at IS NULL);


--
-- Name: idx_calcutta_invitations_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_invitations_active ON core.calcutta_invitations USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_calcutta_invitations_calcutta_id_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_invitations_calcutta_id_active ON core.calcutta_invitations USING btree (calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_calcutta_invitations_invited_by; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_invitations_invited_by ON core.calcutta_invitations USING btree (invited_by);


--
-- Name: idx_calcutta_invitations_user_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_invitations_user_id ON core.calcutta_invitations USING btree (user_id);


--
-- Name: idx_calcutta_scoring_rules_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_scoring_rules_active ON core.calcutta_scoring_rules USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_calcutta_snapshot_entries_entry_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_snapshot_entries_entry_id ON core.calcutta_snapshot_entries USING btree (entry_id);


--
-- Name: idx_calcutta_snapshot_entry_teams_team_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcutta_snapshot_entry_teams_team_id ON core.calcutta_snapshot_entry_teams USING btree (team_id);


--
-- Name: idx_calcuttas_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_calcuttas_active ON core.calcuttas USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_core_calcutta_scoring_rules_calcutta_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id ON core.calcutta_scoring_rules USING btree (calcutta_id);


--
-- Name: idx_core_calcutta_snapshot_entries_snapshot_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_snapshot_entries_snapshot_id ON core.calcutta_snapshot_entries USING btree (calcutta_snapshot_id);


--
-- Name: idx_core_calcutta_snapshot_entry_teams_entry_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_snapshot_entry_teams_entry_id ON core.calcutta_snapshot_entry_teams USING btree (calcutta_snapshot_entry_id);


--
-- Name: idx_core_calcutta_snapshot_payouts_snapshot_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_snapshot_payouts_snapshot_id ON core.calcutta_snapshot_payouts USING btree (calcutta_snapshot_id);


--
-- Name: idx_core_calcutta_snapshot_scoring_rules_snapshot_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_snapshot_scoring_rules_snapshot_id ON core.calcutta_snapshot_scoring_rules USING btree (calcutta_snapshot_id);


--
-- Name: idx_core_calcutta_snapshots_base_calcutta_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcutta_snapshots_base_calcutta_id ON core.calcutta_snapshots USING btree (base_calcutta_id);


--
-- Name: idx_core_calcuttas_budget_points; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcuttas_budget_points ON core.calcuttas USING btree (budget_points);


--
-- Name: idx_core_calcuttas_owner_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcuttas_owner_id ON core.calcuttas USING btree (owner_id);


--
-- Name: idx_core_calcuttas_tournament_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_calcuttas_tournament_id ON core.calcuttas USING btree (tournament_id);


--
-- Name: idx_core_entries_user_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_entries_user_id ON core.entries USING btree (user_id);


--
-- Name: idx_core_payouts_calcutta_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_payouts_calcutta_id ON core.payouts USING btree (calcutta_id);


--
-- Name: idx_core_team_kenpom_stats_team_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_team_kenpom_stats_team_id ON core.team_kenpom_stats USING btree (team_id);


--
-- Name: idx_core_teams_school_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_teams_school_id ON core.teams USING btree (school_id);


--
-- Name: idx_core_teams_tournament_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_teams_tournament_id ON core.teams USING btree (tournament_id);


--
-- Name: idx_core_tournaments_competition_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_tournaments_competition_id ON core.tournaments USING btree (competition_id);


--
-- Name: idx_core_tournaments_season_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_core_tournaments_season_id ON core.tournaments USING btree (season_id);


--
-- Name: idx_entries_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_entries_active ON core.entries USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_entries_calcutta_id_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_entries_calcutta_id_active ON core.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_entry_teams_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_entry_teams_active ON core.entry_teams USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_entry_teams_entry_id_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_entry_teams_entry_id_active ON core.entry_teams USING btree (entry_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_entry_teams_team_id_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_entry_teams_team_id_active ON core.entry_teams USING btree (team_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_grants_label_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_grants_label_id ON core.grants USING btree (label_id);


--
-- Name: idx_grants_permission_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_grants_permission_id ON core.grants USING btree (permission_id);


--
-- Name: idx_grants_revoked_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_grants_revoked_at ON core.grants USING btree (revoked_at);


--
-- Name: idx_grants_scope; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_grants_scope ON core.grants USING btree (scope_type, scope_id);


--
-- Name: idx_grants_user_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_grants_user_id ON core.grants USING btree (user_id);


--
-- Name: idx_idempotency_keys_created; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_idempotency_keys_created ON core.idempotency_keys USING btree (created_at);


--
-- Name: idx_label_permissions_label_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_label_permissions_label_id ON core.label_permissions USING btree (label_id);


--
-- Name: idx_label_permissions_permission_id; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_label_permissions_permission_id ON core.label_permissions USING btree (permission_id);


--
-- Name: idx_payouts_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_payouts_active ON core.payouts USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_teams_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_teams_active ON core.teams USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_tournaments_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_tournaments_active ON core.tournaments USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_active; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_users_active ON core.users USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_invite_expires_at; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_users_invite_expires_at ON core.users USING btree (invite_expires_at);


--
-- Name: idx_users_status; Type: INDEX; Schema: core; Owner: -
--

CREATE INDEX idx_users_status ON core.users USING btree (status);


--
-- Name: uq_api_keys_key_hash; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_api_keys_key_hash ON core.api_keys USING btree (key_hash);


--
-- Name: uq_auth_sessions_refresh_token_hash; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_auth_sessions_refresh_token_hash ON core.auth_sessions USING btree (refresh_token_hash);


--
-- Name: uq_bundle_uploads_sha256; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_bundle_uploads_sha256 ON core.bundle_uploads USING btree (sha256) WHERE (deleted_at IS NULL);


--
-- Name: uq_calcutta_invitations_calcutta_user; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_calcutta_invitations_calcutta_user ON core.calcutta_invitations USING btree (calcutta_id, user_id) WHERE (deleted_at IS NULL);


--
-- Name: uq_core_entry_teams_entry_team; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_core_entry_teams_entry_team ON core.entry_teams USING btree (entry_id, team_id) WHERE (deleted_at IS NULL);


--
-- Name: uq_core_schools_name; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_core_schools_name ON core.schools USING btree (name) WHERE (deleted_at IS NULL);


--
-- Name: uq_core_schools_slug; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_core_schools_slug ON core.schools USING btree (slug) WHERE (deleted_at IS NULL);


--
-- Name: uq_core_tournaments_import_key; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_core_tournaments_import_key ON core.tournaments USING btree (import_key) WHERE (deleted_at IS NULL);


--
-- Name: uq_entries_user_calcutta; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_entries_user_calcutta ON core.entries USING btree (user_id, calcutta_id) WHERE ((user_id IS NOT NULL) AND (deleted_at IS NULL));


--
-- Name: uq_users_invite_token_hash; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX uq_users_invite_token_hash ON core.users USING btree (invite_token_hash) WHERE (invite_token_hash IS NOT NULL);


--
-- Name: users_email_unique; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX users_email_unique ON core.users USING btree (email) WHERE ((email IS NOT NULL) AND (deleted_at IS NULL));


--
-- Name: users_stub_name_unique; Type: INDEX; Schema: core; Owner: -
--

CREATE UNIQUE INDEX users_stub_name_unique ON core.users USING btree (first_name, last_name) WHERE ((status = 'stub'::text) AND (deleted_at IS NULL));


--
-- Name: idx_analytics_calcutta_evaluation_runs_batch_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_calcutta_evaluation_runs_batch_id ON derived.calcutta_evaluation_runs USING btree (simulated_tournament_id);


--
-- Name: idx_analytics_calcutta_evaluation_runs_calcutta_snapshot_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_calcutta_evaluation_runs_calcutta_snapshot_id ON derived.calcutta_evaluation_runs USING btree (calcutta_snapshot_id);


--
-- Name: idx_analytics_entry_performance_eval_run_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_entry_performance_eval_run_id ON derived.entry_performance USING btree (calcutta_evaluation_run_id);


--
-- Name: idx_analytics_simulated_tournaments_batch_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_simulated_tournaments_batch_id ON derived.simulated_teams USING btree (simulated_tournament_id);


--
-- Name: idx_analytics_tournament_simulation_batches_snapshot_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_tournament_simulation_batches_snapshot_id ON derived.simulated_tournaments USING btree (simulation_state_id);


--
-- Name: idx_analytics_tournament_simulation_batches_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_tournament_simulation_batches_tournament_id ON derived.simulated_tournaments USING btree (tournament_id);


--
-- Name: idx_analytics_tournament_state_snapshot_teams_snapshot_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_tournament_state_snapshot_teams_snapshot_id ON derived.simulation_state_teams USING btree (simulation_state_id);


--
-- Name: idx_analytics_tournament_state_snapshot_teams_team_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_tournament_state_snapshot_teams_team_id ON derived.simulation_state_teams USING btree (team_id);


--
-- Name: idx_analytics_tournament_state_snapshots_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_analytics_tournament_state_snapshots_tournament_id ON derived.simulation_states USING btree (tournament_id);


--
-- Name: idx_derived_calcutta_evaluation_runs_run_key; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_calcutta_evaluation_runs_run_key ON derived.calcutta_evaluation_runs USING btree (run_key) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_calcutta_evaluation_runs_simulated_calcutta_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_calcutta_evaluation_runs_simulated_calcutta_id ON derived.calcutta_evaluation_runs USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_game_outcome_runs_created_at; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_game_outcome_runs_created_at ON derived.game_outcome_runs USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_game_outcome_runs_prediction_model_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_game_outcome_runs_prediction_model_id ON derived.game_outcome_runs USING btree (prediction_model_id);


--
-- Name: idx_derived_game_outcome_runs_run_key; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_game_outcome_runs_run_key ON derived.game_outcome_runs USING btree (run_key) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_game_outcome_runs_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_game_outcome_runs_tournament_id ON derived.game_outcome_runs USING btree (tournament_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_predicted_game_outcomes_run_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_predicted_game_outcomes_run_id ON derived.predicted_game_outcomes USING btree (run_id);


--
-- Name: idx_derived_predicted_game_outcomes_team1_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_predicted_game_outcomes_team1_id ON derived.predicted_game_outcomes USING btree (team1_id);


--
-- Name: idx_derived_predicted_game_outcomes_team2_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_predicted_game_outcomes_team2_id ON derived.predicted_game_outcomes USING btree (team2_id);


--
-- Name: idx_derived_run_jobs_kind_status_created_at; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_run_jobs_kind_status_created_at ON derived.run_jobs USING btree (run_kind, status, created_at);


--
-- Name: idx_derived_run_jobs_progress_updated_at; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_run_jobs_progress_updated_at ON derived.run_jobs USING btree (progress_updated_at);


--
-- Name: idx_derived_simulated_calcutta_payouts_simulated_calcutta_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_calcutta_payouts_simulated_calcutta_id ON derived.simulated_calcutta_payouts USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_calcutta_scoring_rules_simulated_calcutta; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_calcutta_scoring_rules_simulated_calcutta ON derived.simulated_calcutta_scoring_rules USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_calcuttas_base_calcutta_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_calcuttas_base_calcutta_id ON derived.simulated_calcuttas USING btree (base_calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_calcuttas_created_at; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_calcuttas_created_at ON derived.simulated_calcuttas USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_calcuttas_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_calcuttas_tournament_id ON derived.simulated_calcuttas USING btree (tournament_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_entries_created_at; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_entries_created_at ON derived.simulated_entries USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_entries_simulated_calcutta_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_entries_simulated_calcutta_id ON derived.simulated_entries USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_entries_source_entry_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_entries_source_entry_id ON derived.simulated_entries USING btree (source_entry_id);


--
-- Name: idx_derived_simulated_entry_teams_simulated_entry_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_entry_teams_simulated_entry_id ON derived.simulated_entry_teams USING btree (simulated_entry_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_derived_simulated_entry_teams_team_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_derived_simulated_entry_teams_team_id ON derived.simulated_entry_teams USING btree (team_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_gold_entry_performance_run_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_gold_entry_performance_run_id ON derived.entry_performance USING btree (run_id);


--
-- Name: idx_silver_predicted_game_outcomes_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_silver_predicted_game_outcomes_tournament_id ON derived.predicted_game_outcomes USING btree (tournament_id);


--
-- Name: idx_silver_simulated_tournaments_sim_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_silver_simulated_tournaments_sim_id ON derived.simulated_teams USING btree (tournament_id, sim_id);


--
-- Name: idx_silver_simulated_tournaments_team_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_silver_simulated_tournaments_team_id ON derived.simulated_teams USING btree (team_id);


--
-- Name: idx_silver_simulated_tournaments_tournament_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX idx_silver_simulated_tournaments_tournament_id ON derived.simulated_teams USING btree (tournament_id);


--
-- Name: prediction_models_kind_idx; Type: INDEX; Schema: derived; Owner: -
--

CREATE INDEX prediction_models_kind_idx ON derived.prediction_models USING btree (kind) WHERE (deleted_at IS NULL);


--
-- Name: prediction_models_kind_name_uniq; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX prediction_models_kind_name_uniq ON derived.prediction_models USING btree (kind, name) WHERE (deleted_at IS NULL);


--
-- Name: uq_analytics_simulated_tournaments_batch_sim_team; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_analytics_simulated_tournaments_batch_sim_team ON derived.simulated_teams USING btree (simulated_tournament_id, sim_id, team_id) WHERE ((deleted_at IS NULL) AND (simulated_tournament_id IS NOT NULL));


--
-- Name: uq_analytics_simulated_tournaments_legacy_sim_team; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_analytics_simulated_tournaments_legacy_sim_team ON derived.simulated_teams USING btree (tournament_id, sim_id, team_id) WHERE ((deleted_at IS NULL) AND (simulated_tournament_id IS NULL));


--
-- Name: uq_analytics_tournament_simulation_batches_natural_key; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_analytics_tournament_simulation_batches_natural_key ON derived.simulated_tournaments USING btree (tournament_id, simulation_state_id, n_sims, seed, probability_source_key) WHERE (deleted_at IS NULL);


--
-- Name: uq_derived_predicted_game_outcomes_legacy_matchup; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_legacy_matchup ON derived.predicted_game_outcomes USING btree (tournament_id, game_id, team1_id, team2_id) WHERE ((run_id IS NULL) AND (deleted_at IS NULL));


--
-- Name: uq_derived_predicted_game_outcomes_run_matchup; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_run_matchup ON derived.predicted_game_outcomes USING btree (run_id, game_id, team1_id, team2_id) WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));


--
-- Name: uq_derived_run_jobs_kind_run_id; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_run_jobs_kind_run_id ON derived.run_jobs USING btree (run_kind, run_id);


--
-- Name: uq_derived_simulated_calcutta_payouts_simulated_position; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_simulated_calcutta_payouts_simulated_position ON derived.simulated_calcutta_payouts USING btree (simulated_calcutta_id, "position") WHERE (deleted_at IS NULL);


--
-- Name: uq_derived_simulated_calcutta_scoring_rules_simulated_win_index; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_simulated_calcutta_scoring_rules_simulated_win_index ON derived.simulated_calcutta_scoring_rules USING btree (simulated_calcutta_id, win_index) WHERE (deleted_at IS NULL);


--
-- Name: uq_derived_simulated_entry_teams_entry_team; Type: INDEX; Schema: derived; Owner: -
--

CREATE UNIQUE INDEX uq_derived_simulated_entry_teams_entry_team ON derived.simulated_entry_teams USING btree (simulated_entry_id, team_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_evaluation_entry_results_evaluation_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_evaluation_entry_results_evaluation_id ON lab.evaluation_entry_results USING btree (evaluation_id);


--
-- Name: idx_lab_entries_calcutta_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_entries_calcutta_id ON lab.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_entries_created_at; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_entries_created_at ON lab.entries USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_entries_investment_model_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_entries_investment_model_id ON lab.entries USING btree (investment_model_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_evaluations_created_at; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_evaluations_created_at ON lab.evaluations USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_evaluations_entry_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_evaluations_entry_id ON lab.evaluations USING btree (entry_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_evaluations_simulated_calcutta_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_evaluations_simulated_calcutta_id ON lab.evaluations USING btree (simulated_calcutta_id);


--
-- Name: idx_lab_investment_models_created_at; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_investment_models_created_at ON lab.investment_models USING btree (created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_investment_models_kind; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_investment_models_kind ON lab.investment_models USING btree (kind) WHERE (deleted_at IS NULL);


--
-- Name: idx_lab_pipeline_calcutta_runs_calcutta_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_calcutta_runs_calcutta_id ON lab.pipeline_calcutta_runs USING btree (calcutta_id);


--
-- Name: idx_lab_pipeline_calcutta_runs_entry_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_calcutta_runs_entry_id ON lab.pipeline_calcutta_runs USING btree (entry_id);


--
-- Name: idx_lab_pipeline_calcutta_runs_evaluation_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_calcutta_runs_evaluation_id ON lab.pipeline_calcutta_runs USING btree (evaluation_id);


--
-- Name: idx_lab_pipeline_calcutta_runs_pipeline_run_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_calcutta_runs_pipeline_run_id ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id);


--
-- Name: idx_lab_pipeline_calcutta_runs_status; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_calcutta_runs_status ON lab.pipeline_calcutta_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));


--
-- Name: idx_lab_pipeline_runs_created_at; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_runs_created_at ON lab.pipeline_runs USING btree (created_at DESC);


--
-- Name: idx_lab_pipeline_runs_investment_model_id; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_runs_investment_model_id ON lab.pipeline_runs USING btree (investment_model_id);


--
-- Name: idx_lab_pipeline_runs_status; Type: INDEX; Schema: lab; Owner: -
--

CREATE INDEX idx_lab_pipeline_runs_status ON lab.pipeline_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));


--
-- Name: uq_lab_entries_model_calcutta_state; Type: INDEX; Schema: lab; Owner: -
--

CREATE UNIQUE INDEX uq_lab_entries_model_calcutta_state ON lab.entries USING btree (investment_model_id, calcutta_id, starting_state_key) WHERE (deleted_at IS NULL);


--
-- Name: uq_lab_evaluations_entry_sims_seed; Type: INDEX; Schema: lab; Owner: -
--

CREATE UNIQUE INDEX uq_lab_evaluations_entry_sims_seed ON lab.evaluations USING btree (entry_id, n_sims, seed) WHERE (deleted_at IS NULL);


--
-- Name: uq_lab_investment_models_name; Type: INDEX; Schema: lab; Owner: -
--

CREATE UNIQUE INDEX uq_lab_investment_models_name ON lab.investment_models USING btree (name) WHERE (deleted_at IS NULL);


--
-- Name: uq_lab_pipeline_calcutta_runs_pipeline_calcutta; Type: INDEX; Schema: lab; Owner: -
--

CREATE UNIQUE INDEX uq_lab_pipeline_calcutta_runs_pipeline_calcutta ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id, calcutta_id);


--
-- Name: api_keys trg_core_api_keys_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_api_keys_updated_at BEFORE UPDATE ON core.api_keys FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: auth_sessions trg_core_auth_sessions_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_auth_sessions_updated_at BEFORE UPDATE ON core.auth_sessions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: bundle_uploads trg_core_bundle_uploads_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_bundle_uploads_updated_at BEFORE UPDATE ON core.bundle_uploads FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_invitations trg_core_calcutta_invitations_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_invitations_updated_at BEFORE UPDATE ON core.calcutta_invitations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_scoring_rules trg_core_calcutta_scoring_rules_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_snapshot_entries trg_core_calcutta_snapshot_entries_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_snapshot_entries_updated_at BEFORE UPDATE ON core.calcutta_snapshot_entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_snapshot_entry_teams trg_core_calcutta_snapshot_entry_teams_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_snapshot_entry_teams_updated_at BEFORE UPDATE ON core.calcutta_snapshot_entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_snapshot_payouts trg_core_calcutta_snapshot_payouts_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_snapshot_payouts_updated_at BEFORE UPDATE ON core.calcutta_snapshot_payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_snapshot_scoring_rules trg_core_calcutta_snapshot_scoring_rules_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_snapshot_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_snapshot_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_snapshots trg_core_calcutta_snapshots_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcutta_snapshots_updated_at BEFORE UPDATE ON core.calcutta_snapshots FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcuttas trg_core_calcuttas_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_calcuttas_updated_at BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: competitions trg_core_competitions_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_competitions_updated_at BEFORE UPDATE ON core.competitions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: entries trg_core_entries_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_entries_updated_at BEFORE UPDATE ON core.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: entry_teams trg_core_entry_teams_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_entry_teams_updated_at BEFORE UPDATE ON core.entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: grants trg_core_grants_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_grants_updated_at BEFORE UPDATE ON core.grants FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: label_permissions trg_core_label_permissions_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_label_permissions_updated_at BEFORE UPDATE ON core.label_permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: labels trg_core_labels_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_labels_updated_at BEFORE UPDATE ON core.labels FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: payouts trg_core_payouts_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_payouts_updated_at BEFORE UPDATE ON core.payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: permissions trg_core_permissions_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_permissions_updated_at BEFORE UPDATE ON core.permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: schools trg_core_schools_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_schools_updated_at BEFORE UPDATE ON core.schools FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: seasons trg_core_seasons_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_seasons_updated_at BEFORE UPDATE ON core.seasons FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: team_kenpom_stats trg_core_team_kenpom_stats_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_team_kenpom_stats_updated_at BEFORE UPDATE ON core.team_kenpom_stats FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: teams trg_core_teams_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_teams_updated_at BEFORE UPDATE ON core.teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: tournaments trg_core_tournaments_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_tournaments_updated_at BEFORE UPDATE ON core.tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: users trg_core_users_updated_at; Type: TRIGGER; Schema: core; Owner: -
--

CREATE TRIGGER trg_core_users_updated_at BEFORE UPDATE ON core.users FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: calcutta_evaluation_runs trg_derived_calcutta_evaluation_runs_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_calcutta_evaluation_runs_updated_at BEFORE UPDATE ON derived.calcutta_evaluation_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: entry_performance trg_derived_entry_performance_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_entry_performance_updated_at BEFORE UPDATE ON derived.entry_performance FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: game_outcome_runs trg_derived_game_outcome_runs_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_game_outcome_runs_updated_at BEFORE UPDATE ON derived.game_outcome_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: predicted_game_outcomes trg_derived_predicted_game_outcomes_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_predicted_game_outcomes_updated_at BEFORE UPDATE ON derived.predicted_game_outcomes FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: prediction_models trg_derived_prediction_models_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_prediction_models_updated_at BEFORE UPDATE ON derived.prediction_models FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: run_jobs trg_derived_run_jobs_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_run_jobs_updated_at BEFORE UPDATE ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_calcutta_payouts trg_derived_simulated_calcutta_payouts_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_calcutta_payouts_updated_at BEFORE UPDATE ON derived.simulated_calcutta_payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_calcutta_scoring_rules trg_derived_simulated_calcutta_scoring_rules_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_calcutta_scoring_rules_updated_at BEFORE UPDATE ON derived.simulated_calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_calcuttas trg_derived_simulated_calcuttas_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_calcuttas_updated_at BEFORE UPDATE ON derived.simulated_calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_entries trg_derived_simulated_entries_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_entries_updated_at BEFORE UPDATE ON derived.simulated_entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_entry_teams trg_derived_simulated_entry_teams_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_entry_teams_updated_at BEFORE UPDATE ON derived.simulated_entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_teams trg_derived_simulated_teams_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_teams_updated_at BEFORE UPDATE ON derived.simulated_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulated_tournaments trg_derived_simulated_tournaments_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulated_tournaments_updated_at BEFORE UPDATE ON derived.simulated_tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulation_state_teams trg_derived_simulation_state_teams_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulation_state_teams_updated_at BEFORE UPDATE ON derived.simulation_state_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: simulation_states trg_derived_simulation_states_updated_at; Type: TRIGGER; Schema: derived; Owner: -
--

CREATE TRIGGER trg_derived_simulation_states_updated_at BEFORE UPDATE ON derived.simulation_states FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: entries trg_lab_entries_updated_at; Type: TRIGGER; Schema: lab; Owner: -
--

CREATE TRIGGER trg_lab_entries_updated_at BEFORE UPDATE ON lab.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: evaluations trg_lab_evaluations_updated_at; Type: TRIGGER; Schema: lab; Owner: -
--

CREATE TRIGGER trg_lab_evaluations_updated_at BEFORE UPDATE ON lab.evaluations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: investment_models trg_lab_investment_models_updated_at; Type: TRIGGER; Schema: lab; Owner: -
--

CREATE TRIGGER trg_lab_investment_models_updated_at BEFORE UPDATE ON lab.investment_models FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: pipeline_calcutta_runs trg_lab_pipeline_calcutta_runs_updated_at; Type: TRIGGER; Schema: lab; Owner: -
--

CREATE TRIGGER trg_lab_pipeline_calcutta_runs_updated_at BEFORE UPDATE ON lab.pipeline_calcutta_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: pipeline_runs trg_lab_pipeline_runs_updated_at; Type: TRIGGER; Schema: lab; Owner: -
--

CREATE TRIGGER trg_lab_pipeline_runs_updated_at BEFORE UPDATE ON lab.pipeline_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();


--
-- Name: api_keys api_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.api_keys
    ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);


--
-- Name: auth_sessions auth_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.auth_sessions
    ADD CONSTRAINT auth_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);


--
-- Name: calcutta_invitations calcutta_invitations_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_invitations
    ADD CONSTRAINT calcutta_invitations_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: calcutta_invitations calcutta_invitations_invited_by_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_invitations
    ADD CONSTRAINT calcutta_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES core.users(id);


--
-- Name: calcutta_invitations calcutta_invitations_user_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_invitations
    ADD CONSTRAINT calcutta_invitations_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);


--
-- Name: calcutta_scoring_rules calcutta_scoring_rules_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_scoring_rules
    ADD CONSTRAINT calcutta_scoring_rules_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;


--
-- Name: calcutta_snapshot_entries calcutta_snapshot_entries_calcutta_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entries
    ADD CONSTRAINT calcutta_snapshot_entries_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;


--
-- Name: calcutta_snapshot_entries calcutta_snapshot_entries_entry_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entries
    ADD CONSTRAINT calcutta_snapshot_entries_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id);


--
-- Name: calcutta_snapshot_entry_teams calcutta_snapshot_entry_teams_calcutta_snapshot_entry_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entry_teams
    ADD CONSTRAINT calcutta_snapshot_entry_teams_calcutta_snapshot_entry_id_fkey FOREIGN KEY (calcutta_snapshot_entry_id) REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE;


--
-- Name: calcutta_snapshot_entry_teams calcutta_snapshot_entry_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_entry_teams
    ADD CONSTRAINT calcutta_snapshot_entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);


--
-- Name: calcutta_snapshot_payouts calcutta_snapshot_payouts_calcutta_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_payouts
    ADD CONSTRAINT calcutta_snapshot_payouts_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;


--
-- Name: calcutta_snapshot_scoring_rules calcutta_snapshot_scoring_rules_calcutta_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules
    ADD CONSTRAINT calcutta_snapshot_scoring_rules_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;


--
-- Name: calcutta_snapshots calcutta_snapshots_base_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcutta_snapshots
    ADD CONSTRAINT calcutta_snapshots_base_calcutta_id_fkey FOREIGN KEY (base_calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: calcuttas calcuttas_owner_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcuttas
    ADD CONSTRAINT calcuttas_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES core.users(id);


--
-- Name: calcuttas calcuttas_tournament_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.calcuttas
    ADD CONSTRAINT calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: entries entries_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entries
    ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: entries entries_user_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entries
    ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);


--
-- Name: entry_teams entry_teams_entry_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entry_teams
    ADD CONSTRAINT entry_teams_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id) ON DELETE CASCADE;


--
-- Name: entry_teams entry_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.entry_teams
    ADD CONSTRAINT entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);


--
-- Name: grants grants_label_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.grants
    ADD CONSTRAINT grants_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);


--
-- Name: grants grants_permission_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.grants
    ADD CONSTRAINT grants_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);


--
-- Name: grants grants_user_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.grants
    ADD CONSTRAINT grants_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);


--
-- Name: label_permissions label_permissions_label_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.label_permissions
    ADD CONSTRAINT label_permissions_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);


--
-- Name: label_permissions label_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.label_permissions
    ADD CONSTRAINT label_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);


--
-- Name: payouts payouts_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.payouts
    ADD CONSTRAINT payouts_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;


--
-- Name: team_kenpom_stats team_kenpom_stats_team_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.team_kenpom_stats
    ADD CONSTRAINT team_kenpom_stats_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;


--
-- Name: teams teams_school_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.teams
    ADD CONSTRAINT teams_school_id_fkey FOREIGN KEY (school_id) REFERENCES core.schools(id);


--
-- Name: teams teams_tournament_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.teams
    ADD CONSTRAINT teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: tournaments tournaments_competition_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.tournaments
    ADD CONSTRAINT tournaments_competition_id_fkey FOREIGN KEY (competition_id) REFERENCES core.competitions(id);


--
-- Name: tournaments tournaments_season_id_fkey; Type: FK CONSTRAINT; Schema: core; Owner: -
--

ALTER TABLE ONLY core.tournaments
    ADD CONSTRAINT tournaments_season_id_fkey FOREIGN KEY (season_id) REFERENCES core.seasons(id);


--
-- Name: calcutta_evaluation_runs calcutta_evaluation_runs_tournament_simulation_batch_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.calcutta_evaluation_runs
    ADD CONSTRAINT calcutta_evaluation_runs_tournament_simulation_batch_id_fkey FOREIGN KEY (simulated_tournament_id) REFERENCES derived.simulated_tournaments(id);


--
-- Name: entry_performance entry_performance_calcutta_evaluation_run_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.entry_performance
    ADD CONSTRAINT entry_performance_calcutta_evaluation_run_id_fkey FOREIGN KEY (calcutta_evaluation_run_id) REFERENCES derived.calcutta_evaluation_runs(id);


--
-- Name: calcutta_evaluation_runs fk_analytics_calcutta_evaluation_runs_calcutta_snapshot_id; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.calcutta_evaluation_runs
    ADD CONSTRAINT fk_analytics_calcutta_evaluation_runs_calcutta_snapshot_id FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id);


--
-- Name: calcutta_evaluation_runs fk_derived_calcutta_evaluation_runs_simulated_calcutta_id; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.calcutta_evaluation_runs
    ADD CONSTRAINT fk_derived_calcutta_evaluation_runs_simulated_calcutta_id FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);


--
-- Name: game_outcome_runs game_outcome_runs_prediction_model_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_prediction_model_id_fkey FOREIGN KEY (prediction_model_id) REFERENCES derived.prediction_models(id);


--
-- Name: game_outcome_runs game_outcome_runs_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.game_outcome_runs
    ADD CONSTRAINT game_outcome_runs_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: predicted_game_outcomes predicted_game_outcomes_run_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_run_id_fkey FOREIGN KEY (run_id) REFERENCES derived.game_outcome_runs(id);


--
-- Name: predicted_game_outcomes predicted_game_outcomes_team1_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE;


--
-- Name: predicted_game_outcomes predicted_game_outcomes_team2_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;


--
-- Name: predicted_game_outcomes predicted_game_outcomes_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;


--
-- Name: simulated_calcutta_payouts simulated_calcutta_payouts_simulated_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcutta_payouts
    ADD CONSTRAINT simulated_calcutta_payouts_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;


--
-- Name: simulated_calcutta_scoring_rules simulated_calcutta_scoring_rules_simulated_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcutta_scoring_rules
    ADD CONSTRAINT simulated_calcutta_scoring_rules_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;


--
-- Name: simulated_calcuttas simulated_calcuttas_base_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcuttas
    ADD CONSTRAINT simulated_calcuttas_base_calcutta_id_fkey FOREIGN KEY (base_calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: simulated_calcuttas simulated_calcuttas_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_calcuttas
    ADD CONSTRAINT simulated_calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: simulated_entries simulated_entries_simulated_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entries
    ADD CONSTRAINT simulated_entries_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;


--
-- Name: simulated_entries simulated_entries_source_entry_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entries
    ADD CONSTRAINT simulated_entries_source_entry_id_fkey FOREIGN KEY (source_entry_id) REFERENCES core.entries(id);


--
-- Name: simulated_entry_teams simulated_entry_teams_simulated_entry_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entry_teams
    ADD CONSTRAINT simulated_entry_teams_simulated_entry_id_fkey FOREIGN KEY (simulated_entry_id) REFERENCES derived.simulated_entries(id) ON DELETE CASCADE;


--
-- Name: simulated_entry_teams simulated_entry_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_entry_teams
    ADD CONSTRAINT simulated_entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);


--
-- Name: simulated_teams simulated_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_teams
    ADD CONSTRAINT simulated_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE RESTRICT;


--
-- Name: simulated_teams simulated_teams_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_teams
    ADD CONSTRAINT simulated_teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE RESTRICT;


--
-- Name: simulated_teams simulated_tournaments_tournament_simulation_batch_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_teams
    ADD CONSTRAINT simulated_tournaments_tournament_simulation_batch_id_fkey FOREIGN KEY (simulated_tournament_id) REFERENCES derived.simulated_tournaments(id);


--
-- Name: simulated_tournaments tournament_simulation_batches_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_tournaments
    ADD CONSTRAINT tournament_simulation_batches_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: simulated_tournaments tournament_simulation_batches_tournament_state_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulated_tournaments
    ADD CONSTRAINT tournament_simulation_batches_tournament_state_snapshot_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id);


--
-- Name: simulation_state_teams tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_state_teams
    ADD CONSTRAINT tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id) ON DELETE CASCADE;


--
-- Name: simulation_state_teams tournament_state_snapshot_teams_team_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_state_teams
    ADD CONSTRAINT tournament_state_snapshot_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);


--
-- Name: simulation_states tournament_state_snapshots_tournament_id_fkey; Type: FK CONSTRAINT; Schema: derived; Owner: -
--

ALTER TABLE ONLY derived.simulation_states
    ADD CONSTRAINT tournament_state_snapshots_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);


--
-- Name: entries entries_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.entries
    ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: entries entries_investment_model_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.entries
    ADD CONSTRAINT entries_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);


--
-- Name: evaluation_entry_results evaluation_entry_results_evaluation_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluation_entry_results
    ADD CONSTRAINT evaluation_entry_results_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);


--
-- Name: evaluations evaluations_entry_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluations
    ADD CONSTRAINT evaluations_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);


--
-- Name: evaluations evaluations_simulated_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.evaluations
    ADD CONSTRAINT evaluations_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);


--
-- Name: pipeline_calcutta_runs pipeline_calcutta_runs_calcutta_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_calcutta_runs
    ADD CONSTRAINT pipeline_calcutta_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);


--
-- Name: pipeline_calcutta_runs pipeline_calcutta_runs_entry_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_calcutta_runs
    ADD CONSTRAINT pipeline_calcutta_runs_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);


--
-- Name: pipeline_calcutta_runs pipeline_calcutta_runs_evaluation_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_calcutta_runs
    ADD CONSTRAINT pipeline_calcutta_runs_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);


--
-- Name: pipeline_calcutta_runs pipeline_calcutta_runs_pipeline_run_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_calcutta_runs
    ADD CONSTRAINT pipeline_calcutta_runs_pipeline_run_id_fkey FOREIGN KEY (pipeline_run_id) REFERENCES lab.pipeline_runs(id) ON DELETE CASCADE;


--
-- Name: pipeline_runs pipeline_runs_investment_model_id_fkey; Type: FK CONSTRAINT; Schema: lab; Owner: -
--

ALTER TABLE ONLY lab.pipeline_runs
    ADD CONSTRAINT pipeline_runs_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);


--
-- PostgreSQL database dump complete
--
