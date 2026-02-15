-- Jubilee Baseline Migration
-- Squashes 119 migrations into a single baseline
-- Changes: drop archive/models schemas, move auth tables public->core,
--          standardize triggers to core.set_updated_at(), move calcutta_slugify to core

-- ============================================================================
-- 1. EXTENSIONS
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

-- ============================================================================
-- 2. SCHEMAS
-- ============================================================================

CREATE SCHEMA core;
CREATE SCHEMA derived;
CREATE SCHEMA lab;

-- ============================================================================
-- 3. FUNCTIONS
-- ============================================================================

CREATE FUNCTION core.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION core.calcutta_slugify(input text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT trim(both '-' from regexp_replace(lower(input), '[^a-z0-9]+', '-', 'g'))
$$;

CREATE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    seed_from_sim INTEGER;
    seed_from_params INTEGER;
    seed_value INTEGER;
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    IF COALESCE(NEW.params_json->>'workerize', 'false') <> 'true' THEN
        RETURN NEW;
    END IF;

    seed_from_sim := NULL;
    SELECT st.seed
    INTO seed_from_sim
    FROM derived.simulated_tournaments st
    WHERE st.id = NEW.simulated_tournament_id
        AND st.deleted_at IS NULL
    LIMIT 1;

    seed_from_params := NULL;
    BEGIN
        seed_from_params := NULLIF(COALESCE(NEW.params_json->>'seed', ''), '')::int;
    EXCEPTION WHEN OTHERS THEN
        seed_from_params := NULL;
    END;

    seed_value := COALESCE(seed_from_params, seed_from_sim);

    dataset_refs := jsonb_build_object(
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'calcutta_snapshot_id', NEW.calcutta_snapshot_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'seed', seed_value,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'calcutta_snapshot_id', NEW.calcutta_snapshot_id,
        'purpose', NEW.purpose,
        'git_sha', NEW.git_sha
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        created_at,
        updated_at
    )
    VALUES (
        'calcutta_evaluation',
        NEW.id,
        NEW.run_key,
        'queued',
        ((base_params || COALESCE(NEW.params_json, '{}'::jsonb)) || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_job_for_entry_evaluation_request() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'calcutta_id', NEW.calcutta_id,
        'entry_candidate_id', NEW.entry_candidate_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.request_source, ''), 'db_trigger'),
        'calcutta_id', NEW.calcutta_id,
        'entry_candidate_id', NEW.entry_candidate_id,
        'excluded_entry_name', NEW.excluded_entry_name,
        'starting_state_key', NEW.starting_state_key,
        'n_sims', NEW.n_sims,
        'seed', NEW.seed,
        'experiment_key', NEW.experiment_key,
        'request_source', NEW.request_source
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        claimed_at,
        claimed_by,
        started_at,
        finished_at,
        error_message,
        created_at,
        updated_at
    )
    VALUES (
        'entry_evaluation',
        NEW.id,
        NEW.run_key,
        NEW.status,
        (base_params || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.claimed_at,
        NEW.claimed_by,
        NEW.claimed_at,
        CASE WHEN NEW.status IN ('succeeded', 'failed') THEN NEW.updated_at ELSE NULL END,
        NEW.error_message,
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_job_for_game_outcome_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'tournament_id', NEW.tournament_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'algorithm_id', NEW.algorithm_id,
        'tournament_id', NEW.tournament_id,
        'git_sha', NEW.git_sha
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        created_at,
        updated_at
    )
    VALUES (
        'game_outcome',
        NEW.id,
        NEW.run_key,
        'queued',
        ((base_params || COALESCE(NEW.params_json, '{}'::jsonb)) || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_job_for_market_share_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'calcutta_id', NEW.calcutta_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'algorithm_id', NEW.algorithm_id,
        'calcutta_id', NEW.calcutta_id,
        'git_sha', NEW.git_sha
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        created_at,
        updated_at
    )
    VALUES (
        'market_share',
        NEW.id,
        NEW.run_key,
        'queued',
        ((base_params || COALESCE(NEW.params_json, '{}'::jsonb)) || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_job_for_simulation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
        'market_share_run_id', NEW.market_share_run_id,
        'strategy_generation_run_id', NEW.strategy_generation_run_id,
        'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
    );

    base_params := jsonb_build_object(
        'source', 'db_trigger',
        'optimizer_key', NEW.optimizer_key,
        'n_sims', NEW.n_sims,
        'seed', NEW.seed,
        'starting_state_key', NEW.starting_state_key,
        'excluded_entry_name', NEW.excluded_entry_name,
        'simulated_calcutta_id', NEW.simulated_calcutta_id,
        'cohort_id', NEW.cohort_id,
        'simulation_run_batch_id', NEW.simulation_run_batch_id,
        'calcutta_id', NEW.calcutta_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'game_outcome_spec_json', NEW.game_outcome_spec_json,
        'market_share_run_id', NEW.market_share_run_id,
        'strategy_generation_run_id', NEW.strategy_generation_run_id,
        'calcutta_evaluation_run_id', NEW.calcutta_evaluation_run_id
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        claimed_at,
        claimed_by,
        started_at,
        finished_at,
        error_message,
        created_at,
        updated_at
    )
    VALUES (
        'simulation',
        NEW.id,
        NEW.run_key,
        NEW.status,
        (base_params || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.claimed_at,
        NEW.claimed_by,
        NEW.claimed_at,
        CASE WHEN NEW.status IN ('succeeded', 'failed') THEN NEW.updated_at ELSE NULL END,
        NEW.error_message,
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_job_for_strategy_generation_run() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    dataset_refs JSONB;
    base_params JSONB;
BEGIN
    dataset_refs := jsonb_build_object(
        'calcutta_id', NEW.calcutta_id,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'market_share_run_id', NEW.market_share_run_id,
        'game_outcome_run_id', NEW.game_outcome_run_id
    );

    base_params := jsonb_build_object(
        'source', COALESCE(NULLIF(NEW.params_json->>'source', ''), 'db_trigger'),
        'calcutta_id', NEW.calcutta_id,
        'simulated_tournament_id', NEW.simulated_tournament_id,
        'name', NEW.name,
        'purpose', NEW.purpose,
        'returns_model_key', NEW.returns_model_key,
        'investment_model_key', NEW.investment_model_key,
        'optimizer_key', NEW.optimizer_key,
        'git_sha', NEW.git_sha,
        'market_share_run_id', NEW.market_share_run_id,
        'game_outcome_run_id', NEW.game_outcome_run_id,
        'excluded_entry_name', NEW.excluded_entry_name,
        'starting_state_key', NEW.starting_state_key
    );

    INSERT INTO derived.run_jobs (
        run_kind,
        run_id,
        run_key,
        status,
        params_json,
        created_at,
        updated_at
    )
    VALUES (
        'strategy_generation',
        NEW.id,
        NEW.run_key_uuid,
        'queued',
        ((base_params || COALESCE(NEW.params_json, '{}'::jsonb)) || jsonb_build_object('dataset_refs', dataset_refs)),
        NEW.created_at,
        NEW.updated_at
    )
    ON CONFLICT (run_kind, run_id)
    DO NOTHING;

    RETURN NEW;
END;
$$;

CREATE FUNCTION derived.enqueue_run_progress_event_from_run_jobs() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO derived.run_progress_events (
            run_kind,
            run_id,
            run_key,
            event_kind,
            status,
            percent,
            phase,
            message,
            source,
            payload_json,
            created_at
        )
        VALUES (
            NEW.run_kind,
            NEW.run_id,
            NEW.run_key,
            'status',
            NEW.status,
            NULL,
            NEW.status,
            NULL,
            'db_trigger',
            '{}'::jsonb,
            COALESCE(NEW.created_at, NOW())
        );

        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF NEW.status IS DISTINCT FROM OLD.status THEN
            INSERT INTO derived.run_progress_events (
                run_kind,
                run_id,
                run_key,
                event_kind,
                status,
                percent,
                phase,
                message,
                source,
                payload_json,
                created_at
            )
            VALUES (
                NEW.run_kind,
                NEW.run_id,
                NEW.run_key,
                'status',
                NEW.status,
                NULL,
                NEW.status,
                CASE WHEN NEW.status = 'failed' THEN NEW.error_message ELSE NULL END,
                'db_trigger',
                '{}'::jsonb,
                COALESCE(NEW.finished_at, NEW.updated_at, NOW())
            );
        END IF;

        RETURN NEW;
    END IF;

    RETURN NEW;
END;
$$;

-- ============================================================================
-- 4. CORE AUTH TABLES (moved from public)
-- ============================================================================

CREATE TABLE core.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255) NOT NULL,
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
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'invited'::text, 'requires_password_setup'::text])))
);

CREATE TABLE core.permissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.labels (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.label_permissions (
    label_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

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
    CONSTRAINT grants_scope_type_check CHECK ((scope_type = ANY (ARRAY['global'::text, 'calcutta'::text]))),
    CONSTRAINT grants_subject_check CHECK ((((label_id IS NOT NULL) AND (permission_id IS NULL)) OR ((label_id IS NULL) AND (permission_id IS NOT NULL))))
);

-- ============================================================================
-- 5. CORE DOMAIN TABLES
-- ============================================================================

CREATE TABLE core.seasons (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    year integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.competitions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.schools (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

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

CREATE TABLE core.teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    school_id uuid NOT NULL,
    seed integer NOT NULL,
    region character varying(50) NOT NULL,
    byes integer NOT NULL DEFAULT 0,
    wins integer NOT NULL DEFAULT 0,
    eliminated boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

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

CREATE TABLE core.calcuttas (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    min_teams integer NOT NULL DEFAULT 3,
    max_teams integer NOT NULL DEFAULT 10,
    max_bid integer NOT NULL DEFAULT 50,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    budget_points integer NOT NULL DEFAULT 100
);

CREATE TABLE core.entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    user_id uuid,
    calcutta_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_snapshots (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    base_calcutta_id uuid NOT NULL,
    snapshot_type text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_snapshot_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    entry_id uuid,
    display_name text NOT NULL,
    is_synthetic boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_snapshot_entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_snapshot_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_snapshot_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_snapshot_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE core.calcutta_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- Function that depends on core.calcutta_scoring_rules (must come after table creation)
CREATE FUNCTION core.calcutta_points_for_progress(p_calcutta_id uuid, p_wins integer, p_byes integer DEFAULT 0) RETURNS integer
    LANGUAGE sql STABLE
    AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.calcutta_scoring_rules r
    WHERE r.calcutta_id = p_calcutta_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;

-- ============================================================================
-- 6. DERIVED TABLES
-- ============================================================================

CREATE TABLE derived.simulation_states (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    source text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.simulation_state_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulation_state_id uuid NOT NULL,
    team_id uuid NOT NULL,
    wins integer NOT NULL,
    byes integer NOT NULL,
    eliminated boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

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

CREATE TABLE derived.simulated_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    sim_id integer NOT NULL,
    team_id uuid NOT NULL,
    wins integer NOT NULL,
    byes integer NOT NULL DEFAULT 0,
    eliminated boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    simulated_tournament_id uuid
);

CREATE TABLE derived.prediction_models (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    kind text NOT NULL,
    name text NOT NULL,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.game_outcome_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    algorithm_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    git_sha text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_key uuid NOT NULL DEFAULT public.uuid_generate_v4(),
    prediction_model_id uuid
);

CREATE TABLE derived.calcutta_evaluation_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
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

CREATE TABLE derived.entry_simulation_outcomes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_id character varying(255) NOT NULL,
    entry_name character varying(255) NOT NULL,
    sim_id integer NOT NULL,
    payout_cents integer NOT NULL,
    rank integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    points_scored double precision NOT NULL DEFAULT 0,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    calcutta_evaluation_run_id uuid
);

CREATE TABLE derived.entry_bids (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid NOT NULL,
    entry_name character varying(255) NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.detailed_investment_report (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_id character varying(255) NOT NULL,
    team_id uuid NOT NULL,
    expected_points double precision NOT NULL,
    predicted_market_points double precision NOT NULL,
    actual_market_points double precision,
    our_bid_points integer,
    expected_roi double precision NOT NULL,
    our_roi double precision,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    strategy_generation_run_id uuid
);

CREATE TABLE derived.optimization_runs (
    run_id character varying(255) NOT NULL,
    calcutta_id uuid,
    strategy character varying(100) NOT NULL,
    n_sims integer NOT NULL,
    seed integer NOT NULL,
    budget_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    strategy_generation_run_id uuid
);

CREATE TABLE derived.predicted_game_outcomes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tournament_id uuid NOT NULL,
    game_id character varying(255) NOT NULL,
    round integer NOT NULL,
    team1_id uuid NOT NULL,
    team2_id uuid NOT NULL,
    p_team1_wins double precision NOT NULL,
    p_matchup double precision NOT NULL DEFAULT 1.0,
    model_version character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_id uuid
);

CREATE TABLE derived.predicted_market_share (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    calcutta_id uuid,
    team_id uuid NOT NULL,
    predicted_share double precision NOT NULL,
    predicted_points double precision NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tournament_id uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    run_id uuid,
    CONSTRAINT market_share_must_have_calcutta_or_tournament CHECK (((calcutta_id IS NOT NULL) OR (tournament_id IS NOT NULL)))
);

CREATE TABLE derived.run_artifacts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid,
    artifact_kind text NOT NULL,
    schema_version text NOT NULL,
    storage_uri text,
    summary_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    input_market_share_artifact_id uuid,
    input_advancement_artifact_id uuid,
    CONSTRAINT ck_derived_run_artifacts_strategy_generation_lineage CHECK (((run_kind <> 'strategy_generation'::text) OR (artifact_kind <> 'metrics'::text) OR ((input_market_share_artifact_id IS NOT NULL) <> (input_advancement_artifact_id IS NOT NULL))))
);

CREATE TABLE derived.run_jobs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid NOT NULL,
    status text NOT NULL DEFAULT 'queued'::text,
    attempt integer NOT NULL DEFAULT 0,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    progress_json jsonb NOT NULL DEFAULT '{}'::jsonb,
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

CREATE TABLE derived.run_progress_events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid,
    event_kind text NOT NULL,
    status text,
    percent double precision,
    phase text,
    message text,
    source text NOT NULL DEFAULT 'unknown'::text,
    payload_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_derived_run_progress_events_event_kind CHECK ((event_kind = ANY (ARRAY['status'::text, 'progress'::text]))),
    CONSTRAINT ck_derived_run_progress_events_percent CHECK (((percent IS NULL) OR ((percent >= (0)::double precision) AND (percent <= (1)::double precision)))),
    CONSTRAINT ck_derived_run_progress_events_status CHECK (((status IS NULL) OR (status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text]))))
);

CREATE TABLE derived.simulated_calcuttas (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    tournament_id uuid NOT NULL,
    base_calcutta_id uuid,
    starting_state_key text NOT NULL DEFAULT 'post_first_four'::text,
    excluded_entry_name text,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    highlighted_simulated_entry_id uuid,
    CONSTRAINT ck_derived_simulated_calcuttas_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text])))
);

CREATE TABLE derived.simulated_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    display_name text NOT NULL,
    source_kind text NOT NULL,
    source_entry_id uuid,
    source_candidate_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_simulated_entries_source_kind CHECK ((source_kind = ANY (ARRAY['manual'::text, 'from_real_entry'::text, 'from_candidate'::text])))
);

CREATE TABLE derived.simulated_entry_teams (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_entry_id uuid NOT NULL,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.simulated_calcutta_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    "position" integer NOT NULL,
    amount_cents integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.simulated_calcutta_scoring_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    simulated_calcutta_id uuid NOT NULL,
    win_index integer NOT NULL,
    points_awarded integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE derived.simulation_cohorts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    game_outcomes_algorithm_id uuid NOT NULL,
    market_share_algorithm_id uuid NOT NULL,
    optimizer_key text NOT NULL,
    n_sims integer NOT NULL,
    seed integer NOT NULL,
    starting_state_key text NOT NULL DEFAULT 'post_first_four'::text,
    excluded_entry_name text,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_n_sims CHECK ((n_sims > 0)),
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_seed CHECK ((seed <> 0)),
    CONSTRAINT ck_derived_synthetic_calcutta_cohorts_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text])))
);

CREATE TABLE derived.simulation_run_batches (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    cohort_id uuid NOT NULL,
    name text,
    optimizer_key text,
    n_sims integer,
    seed integer,
    starting_state_key text NOT NULL DEFAULT 'post_first_four'::text,
    excluded_entry_name text,
    status text NOT NULL DEFAULT 'running'::text,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT ck_derived_simulation_run_batches_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text]))),
    CONSTRAINT ck_derived_simulation_run_batches_status CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

CREATE TABLE derived.simulation_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_key uuid NOT NULL DEFAULT public.uuid_generate_v4(),
    simulation_run_batch_id uuid,
    cohort_id uuid NOT NULL,
    calcutta_id uuid,
    game_outcome_run_id uuid,
    market_share_run_id uuid,
    strategy_generation_run_id uuid,
    calcutta_evaluation_run_id uuid,
    starting_state_key text NOT NULL DEFAULT 'post_first_four'::text,
    excluded_entry_name text,
    optimizer_key text,
    n_sims integer,
    seed integer,
    our_rank integer,
    our_mean_normalized_payout double precision,
    our_median_normalized_payout double precision,
    our_p_top1 double precision,
    our_p_in_money double precision,
    total_simulations integer,
    realized_finish_position integer,
    realized_is_tied boolean,
    realized_in_the_money boolean,
    realized_payout_cents integer,
    realized_total_points double precision,
    status text NOT NULL DEFAULT 'queued'::text,
    claimed_at timestamp with time zone,
    claimed_by text,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    focus_snapshot_entry_id uuid,
    simulated_calcutta_id uuid,
    game_outcome_spec_json jsonb,
    CONSTRAINT ck_derived_simulation_runs_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['post_first_four'::text, 'current'::text]))),
    CONSTRAINT ck_derived_simulation_runs_status CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

CREATE TABLE derived.optimized_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_key text,
    name text,
    calcutta_id uuid NOT NULL,
    simulated_tournament_id uuid,
    game_outcome_run_id uuid,
    market_share_run_id uuid,
    optimizer_kind text NOT NULL DEFAULT 'minlp'::text,
    optimizer_params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    bids_json jsonb NOT NULL,
    purpose text,
    excluded_entry_name text,
    starting_state_key text,
    returns_model_key text,
    investment_model_key text,
    git_sha text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- ============================================================================
-- 7. LAB TABLES
-- ============================================================================

CREATE TABLE lab.investment_models (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    kind text NOT NULL,
    params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    is_benchmark boolean NOT NULL DEFAULT false
);

COMMENT ON COLUMN lab.investment_models.is_benchmark IS 'True for oracle/baseline models that should be excluded from training data';

CREATE TABLE lab.entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    investment_model_id uuid NOT NULL,
    calcutta_id uuid NOT NULL,
    game_outcome_kind text NOT NULL DEFAULT 'kenpom'::text,
    game_outcome_params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    optimizer_kind text NOT NULL DEFAULT 'minlp'::text,
    optimizer_params_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    starting_state_key text NOT NULL DEFAULT 'post_first_four'::text,
    bids_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    predictions_json jsonb,
    state text NOT NULL DEFAULT 'complete'::text,
    CONSTRAINT ck_lab_entries_starting_state_key CHECK ((starting_state_key = ANY (ARRAY['pre_tournament'::text, 'post_first_four'::text, 'current'::text]))),
    CONSTRAINT ck_lab_entries_state CHECK ((state = ANY (ARRAY['pending_predictions'::text, 'pending_optimization'::text, 'pending_evaluation'::text, 'complete'::text])))
);

COMMENT ON COLUMN lab.entries.bids_json IS 'Optimized bids: [{team_id, bid_points, expected_roi}]. Our optimal allocation given predictions.';
COMMENT ON COLUMN lab.entries.predictions_json IS 'Market predictions: [{team_id, predicted_market_share, expected_points}]. What model predicts others will bid.';

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

CREATE TABLE lab.pipeline_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    investment_model_id uuid NOT NULL,
    target_calcutta_ids uuid[] NOT NULL,
    budget_points integer NOT NULL DEFAULT 10000,
    optimizer_kind text NOT NULL DEFAULT 'predicted_market_share'::text,
    n_sims integer NOT NULL DEFAULT 10000,
    seed integer NOT NULL DEFAULT 42,
    excluded_entry_name text,
    status text NOT NULL DEFAULT 'pending'::text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_lab_pipeline_runs_budget_points CHECK ((budget_points > 0)),
    CONSTRAINT ck_lab_pipeline_runs_n_sims CHECK ((n_sims > 0)),
    CONSTRAINT ck_lab_pipeline_runs_status CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'cancelled'::text])))
);

CREATE TABLE lab.pipeline_calcutta_runs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    pipeline_run_id uuid NOT NULL,
    calcutta_id uuid NOT NULL,
    entry_id uuid,
    stage text NOT NULL DEFAULT 'predictions'::text,
    status text NOT NULL DEFAULT 'pending'::text,
    progress double precision NOT NULL DEFAULT 0.0,
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

-- ============================================================================
-- 8. CORE VIEWS
-- ============================================================================

CREATE VIEW core.derived_portfolio_teams AS
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
        CASE
            WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
            ELSE (0)::double precision
        END AS ownership_percentage,
        (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision AS team_points,
        (core.calcutta_points_for_progress(eb.calcutta_id, eb.tournament_rounds, 0))::double precision AS team_max_points,
        eb.school_id,
        eb.tournament_id,
        eb.seed,
        eb.region,
        eb.byes,
        eb.wins,
        eb.eliminated,
        eb.entry_team_created_at,
        eb.entry_team_updated_at,
        eb.team_created_at,
        eb.team_updated_at,
        eb.school_name,
        eb.derived_updated_at
    FROM entry_bids eb
)
SELECT md5(concat((entry_id)::text, ':', (team_id)::text)) AS id,
    entry_id AS portfolio_id,
    team_id,
    ownership_percentage,
    (team_points * ownership_percentage) AS actual_points,
    CASE
        WHEN eliminated THEN (team_points * ownership_percentage)
        ELSE (team_max_points * ownership_percentage)
    END AS expected_points,
    CASE
        WHEN eliminated THEN (team_points * ownership_percentage)
        ELSE (team_max_points * ownership_percentage)
    END AS predicted_points,
    entry_team_created_at AS created_at,
    derived_updated_at AS updated_at,
    NULL::timestamp with time zone AS deleted_at,
    team_id AS tournament_team_id,
    school_id,
    tournament_id,
    seed,
    region,
    byes,
    wins,
    eliminated,
    team_created_at,
    team_updated_at,
    school_name
FROM entry_team_points etp;

CREATE VIEW core.derived_portfolios AS
WITH entry_totals AS (
    SELECT dpt.portfolio_id AS entry_id,
        sum(dpt.expected_points) AS maximum_points,
        max(dpt.updated_at) AS updated_at
    FROM core.derived_portfolio_teams dpt
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

-- ============================================================================
-- 9. DERIVED VIEWS
-- ============================================================================

CREATE VIEW derived.calcuttas AS
SELECT id,
    id AS core_calcutta_id,
    tournament_id,
    tournament_id AS core_tournament_id,
    owner_id,
    name,
    min_teams,
    max_teams,
    max_bid,
    created_at,
    updated_at,
    deleted_at
FROM core.calcuttas c;

CREATE VIEW derived.teams AS
SELECT tm.id,
    tm.id AS core_team_id,
    tm.tournament_id,
    tm.tournament_id AS core_tournament_id,
    s.name AS school_name,
    tm.seed,
    tm.region,
    ks.net_rtg AS kenpom_net,
    tm.created_at,
    tm.updated_at,
    tm.deleted_at
FROM ((core.teams tm
    LEFT JOIN core.schools s ON ((s.id = tm.school_id)))
    LEFT JOIN core.team_kenpom_stats ks ON ((ks.team_id = tm.id)));

CREATE VIEW derived.tournaments AS
SELECT t.id,
    t.id AS core_tournament_id,
    seas.year AS season,
    t.name,
    t.import_key,
    t.rounds,
    t.starting_at,
    t.final_four_top_left,
    t.final_four_bottom_left,
    t.final_four_top_right,
    t.final_four_bottom_right,
    t.created_at,
    t.updated_at,
    t.deleted_at
FROM (core.tournaments t
    LEFT JOIN core.seasons seas ON ((seas.id = t.season_id)));

CREATE VIEW derived.v_algorithms AS
SELECT id,
    kind,
    name,
    description,
    params_json,
    created_at,
    updated_at,
    deleted_at
FROM derived.prediction_models;

CREATE VIEW derived.v_strategy_generation_run_bids AS
SELECT oe.id AS strategy_generation_run_id,
    oe.run_key AS run_id,
    ((bid.value ->> 'team_id'::text))::uuid AS team_id,
    ((bid.value ->> 'bid_points'::text))::integer AS bid_points,
    COALESCE(((bid.value ->> 'expected_roi'::text))::double precision, (0)::double precision) AS expected_roi,
    oe.created_at,
    oe.deleted_at
FROM derived.optimized_entries oe,
    LATERAL jsonb_array_elements(oe.bids_json) bid(value);

-- ============================================================================
-- 10. LAB VIEWS
-- ============================================================================

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

-- ============================================================================
-- 11. PRIMARY KEY CONSTRAINTS
-- ============================================================================

-- Core auth tables
ALTER TABLE ONLY core.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.permissions ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.labels ADD CONSTRAINT labels_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_pkey PRIMARY KEY (label_id, permission_id);
ALTER TABLE ONLY core.api_keys ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.auth_sessions ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.bundle_uploads ADD CONSTRAINT bundle_uploads_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_pkey PRIMARY KEY (id);

-- Core domain tables
ALTER TABLE ONLY core.seasons ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.competitions ADD CONSTRAINT competitions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.schools ADD CONSTRAINT schools_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.team_kenpom_stats ADD CONSTRAINT team_kenpom_stats_pkey PRIMARY KEY (team_id);
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.payouts ADD CONSTRAINT payouts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_snapshots ADD CONSTRAINT calcutta_snapshots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_snapshot_entries ADD CONSTRAINT calcutta_snapshot_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_snapshot_entry_teams ADD CONSTRAINT calcutta_snapshot_entry_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_snapshot_payouts ADD CONSTRAINT calcutta_snapshot_payouts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules ADD CONSTRAINT calcutta_snapshot_scoring_rules_pkey PRIMARY KEY (id);
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_pkey PRIMARY KEY (id);

-- Derived tables
ALTER TABLE ONLY derived.entry_bids ADD CONSTRAINT bronze_entry_bids_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.calcutta_evaluation_runs ADD CONSTRAINT calcutta_evaluation_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.game_outcome_runs ADD CONSTRAINT game_outcome_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.detailed_investment_report ADD CONSTRAINT gold_detailed_investment_report_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.entry_performance ADD CONSTRAINT gold_entry_performance_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.entry_simulation_outcomes ADD CONSTRAINT gold_entry_simulation_outcomes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.optimization_runs ADD CONSTRAINT gold_optimization_runs_pkey PRIMARY KEY (run_id);
ALTER TABLE ONLY derived.optimized_entries ADD CONSTRAINT optimized_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.prediction_models ADD CONSTRAINT prediction_models_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.run_jobs ADD CONSTRAINT run_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.run_progress_events ADD CONSTRAINT run_progress_events_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT silver_predicted_game_outcomes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT silver_predicted_market_share_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT silver_simulated_tournaments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_calcutta_payouts ADD CONSTRAINT simulated_calcutta_payouts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_calcutta_scoring_rules ADD CONSTRAINT simulated_calcutta_scoring_rules_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_calcuttas ADD CONSTRAINT simulated_calcuttas_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_entries ADD CONSTRAINT simulated_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_entry_teams ADD CONSTRAINT simulated_entry_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_run_batches ADD CONSTRAINT simulation_run_batches_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_cohorts ADD CONSTRAINT synthetic_calcutta_cohorts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT tournament_simulation_batches_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT tournament_state_snapshot_teams_pkey PRIMARY KEY (id);
ALTER TABLE ONLY derived.simulation_states ADD CONSTRAINT tournament_state_snapshots_pkey PRIMARY KEY (id);

-- Lab tables
ALTER TABLE ONLY lab.entries ADD CONSTRAINT lab_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.evaluation_entry_results ADD CONSTRAINT evaluation_entry_results_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.evaluations ADD CONSTRAINT evaluations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.investment_models ADD CONSTRAINT investment_models_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY lab.pipeline_runs ADD CONSTRAINT pipeline_runs_pkey PRIMARY KEY (id);

-- ============================================================================
-- 12. UNIQUE CONSTRAINTS
-- ============================================================================

-- Core auth tables
ALTER TABLE ONLY core.users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY core.labels ADD CONSTRAINT labels_key_key UNIQUE (key);
ALTER TABLE ONLY core.permissions ADD CONSTRAINT permissions_key_key UNIQUE (key);

-- Core domain tables
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT uq_core_calcutta_scoring_rules UNIQUE (calcutta_id, win_index);
ALTER TABLE ONLY core.calcutta_snapshot_entries ADD CONSTRAINT uq_core_calcutta_snapshot_entries_snapshot_display_name UNIQUE (calcutta_snapshot_id, display_name);
ALTER TABLE ONLY core.calcutta_snapshot_entry_teams ADD CONSTRAINT uq_core_calcutta_snapshot_entry_teams_entry_team UNIQUE (calcutta_snapshot_entry_id, team_id);
ALTER TABLE ONLY core.calcutta_snapshot_payouts ADD CONSTRAINT uq_core_calcutta_snapshot_payouts_snapshot_position UNIQUE (calcutta_snapshot_id, "position");
ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules ADD CONSTRAINT uq_core_calcutta_snapshot_scoring_rules_snapshot_win_index UNIQUE (calcutta_snapshot_id, win_index);
ALTER TABLE ONLY core.competitions ADD CONSTRAINT uq_core_competitions_name UNIQUE (name);
ALTER TABLE ONLY core.payouts ADD CONSTRAINT uq_core_payouts_calcutta_position UNIQUE (calcutta_id, "position");
ALTER TABLE ONLY core.seasons ADD CONSTRAINT uq_core_seasons_year UNIQUE (year);

-- Derived tables
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT uq_analytics_tournament_state_snapshot_teams_snapshot_team UNIQUE (simulation_state_id, team_id);

-- Lab tables
ALTER TABLE ONLY lab.evaluation_entry_results ADD CONSTRAINT evaluation_entry_results_evaluation_id_entry_name_key UNIQUE (evaluation_id, entry_name);

-- ============================================================================
-- 13. INDEXES
-- ============================================================================

-- Core auth table indexes (moved from public to core)
CREATE INDEX idx_api_keys_revoked_at ON core.api_keys USING btree (revoked_at);
CREATE INDEX idx_api_keys_user_id ON core.api_keys USING btree (user_id);
CREATE INDEX idx_auth_sessions_expires_at ON core.auth_sessions USING btree (expires_at);
CREATE INDEX idx_auth_sessions_revoked_at ON core.auth_sessions USING btree (revoked_at);
CREATE INDEX idx_auth_sessions_user_id ON core.auth_sessions USING btree (user_id);
CREATE INDEX idx_bundle_uploads_created_at ON core.bundle_uploads USING btree (created_at);
CREATE INDEX idx_bundle_uploads_status_created_at ON core.bundle_uploads USING btree (status, created_at) WHERE (deleted_at IS NULL);
CREATE INDEX idx_grants_revoked_at ON core.grants USING btree (revoked_at);
CREATE INDEX idx_grants_scope ON core.grants USING btree (scope_type, scope_id);
CREATE INDEX idx_grants_user_id ON core.grants USING btree (user_id);
CREATE INDEX idx_users_invite_expires_at ON core.users USING btree (invite_expires_at);
CREATE INDEX idx_users_status ON core.users USING btree (status);
CREATE UNIQUE INDEX uq_api_keys_key_hash ON core.api_keys USING btree (key_hash);
CREATE UNIQUE INDEX uq_auth_sessions_refresh_token_hash ON core.auth_sessions USING btree (refresh_token_hash);
CREATE UNIQUE INDEX uq_bundle_uploads_sha256 ON core.bundle_uploads USING btree (sha256) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_users_invite_token_hash ON core.users USING btree (invite_token_hash) WHERE (invite_token_hash IS NOT NULL);

-- Core domain indexes
CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id ON core.calcutta_scoring_rules USING btree (calcutta_id);
CREATE INDEX idx_core_calcutta_snapshot_entries_snapshot_id ON core.calcutta_snapshot_entries USING btree (calcutta_snapshot_id);
CREATE INDEX idx_core_calcutta_snapshot_entry_teams_entry_id ON core.calcutta_snapshot_entry_teams USING btree (calcutta_snapshot_entry_id);
CREATE INDEX idx_core_calcutta_snapshot_payouts_snapshot_id ON core.calcutta_snapshot_payouts USING btree (calcutta_snapshot_id);
CREATE INDEX idx_core_calcutta_snapshot_scoring_rules_snapshot_id ON core.calcutta_snapshot_scoring_rules USING btree (calcutta_snapshot_id);
CREATE INDEX idx_core_calcutta_snapshots_base_calcutta_id ON core.calcutta_snapshots USING btree (base_calcutta_id);
CREATE INDEX idx_core_calcuttas_budget_points ON core.calcuttas USING btree (budget_points);
CREATE INDEX idx_core_calcuttas_owner_id ON core.calcuttas USING btree (owner_id);
CREATE INDEX idx_core_calcuttas_tournament_id ON core.calcuttas USING btree (tournament_id);
CREATE INDEX idx_core_entries_calcutta_id ON core.entries USING btree (calcutta_id);
CREATE INDEX idx_core_entries_user_id ON core.entries USING btree (user_id);
CREATE INDEX idx_core_entry_teams_entry_id ON core.entry_teams USING btree (entry_id);
CREATE INDEX idx_core_entry_teams_team_id ON core.entry_teams USING btree (team_id);
CREATE INDEX idx_core_payouts_calcutta_id ON core.payouts USING btree (calcutta_id);
CREATE INDEX idx_core_team_kenpom_stats_team_id ON core.team_kenpom_stats USING btree (team_id);
CREATE INDEX idx_core_teams_school_id ON core.teams USING btree (school_id);
CREATE INDEX idx_core_teams_tournament_id ON core.teams USING btree (tournament_id);
CREATE INDEX idx_core_tournaments_competition_id ON core.tournaments USING btree (competition_id);
CREATE INDEX idx_core_tournaments_season_id ON core.tournaments USING btree (season_id);
CREATE UNIQUE INDEX uq_core_schools_name ON core.schools USING btree (name) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_core_schools_slug ON core.schools USING btree (slug) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_core_tournaments_import_key ON core.tournaments USING btree (import_key) WHERE (deleted_at IS NULL);

-- Derived indexes
CREATE INDEX idx_analytics_calcutta_evaluation_runs_batch_id ON derived.calcutta_evaluation_runs USING btree (simulated_tournament_id);
CREATE INDEX idx_analytics_calcutta_evaluation_runs_calcutta_snapshot_id ON derived.calcutta_evaluation_runs USING btree (calcutta_snapshot_id);
CREATE INDEX idx_derived_calcutta_evaluation_runs_run_key ON derived.calcutta_evaluation_runs USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_calcutta_evaluation_runs_simulated_calcutta_id ON derived.calcutta_evaluation_runs USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_gold_detailed_investment_report_run_id ON derived.detailed_investment_report USING btree (run_id);
CREATE INDEX idx_lab_gold_detailed_investment_report_strategy_generation_run ON derived.detailed_investment_report USING btree (strategy_generation_run_id);
CREATE INDEX idx_bronze_entry_bids_calcutta_id ON derived.entry_bids USING btree (calcutta_id);
CREATE INDEX idx_bronze_entry_bids_team_id ON derived.entry_bids USING btree (team_id);
CREATE INDEX idx_analytics_entry_performance_eval_run_id ON derived.entry_performance USING btree (calcutta_evaluation_run_id);
CREATE INDEX idx_gold_entry_performance_run_id ON derived.entry_performance USING btree (run_id);
CREATE INDEX idx_analytics_entry_simulation_outcomes_eval_run_id ON derived.entry_simulation_outcomes USING btree (calcutta_evaluation_run_id);
CREATE INDEX idx_gold_entry_simulation_outcomes_run_id ON derived.entry_simulation_outcomes USING btree (run_id);
CREATE INDEX idx_derived_game_outcome_runs_algorithm_id ON derived.game_outcome_runs USING btree (algorithm_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_game_outcome_runs_created_at ON derived.game_outcome_runs USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_game_outcome_runs_run_key ON derived.game_outcome_runs USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_game_outcome_runs_tournament_id ON derived.game_outcome_runs USING btree (tournament_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_gold_optimization_runs_strategy_generation_run_id ON derived.optimization_runs USING btree (strategy_generation_run_id);
CREATE INDEX optimized_entries_calcutta_id_idx ON derived.optimized_entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX optimized_entries_game_outcome_run_id_idx ON derived.optimized_entries USING btree (game_outcome_run_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX optimized_entries_run_key_uniq ON derived.optimized_entries USING btree (run_key) WHERE ((run_key IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_silver_predicted_game_outcomes_tournament_id ON derived.predicted_game_outcomes USING btree (tournament_id);
CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_legacy_matchup ON derived.predicted_game_outcomes USING btree (tournament_id, game_id, team1_id, team2_id) WHERE ((run_id IS NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_derived_predicted_game_outcomes_run_matchup ON derived.predicted_game_outcomes USING btree (run_id, game_id, team1_id, team2_id) WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_derived_predicted_market_share_legacy_calcutta_team ON derived.predicted_market_share USING btree (calcutta_id, team_id) WHERE ((calcutta_id IS NOT NULL) AND (run_id IS NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_derived_predicted_market_share_legacy_tournament_team ON derived.predicted_market_share USING btree (tournament_id, team_id) WHERE ((tournament_id IS NOT NULL) AND (run_id IS NULL) AND (deleted_at IS NULL));
CREATE UNIQUE INDEX uq_derived_predicted_market_share_run_team ON derived.predicted_market_share USING btree (run_id, team_id) WHERE ((run_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX prediction_models_kind_idx ON derived.prediction_models USING btree (kind) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX prediction_models_kind_name_uniq ON derived.prediction_models USING btree (kind, name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_input_advancement_artifact_id ON derived.run_artifacts USING btree (input_advancement_artifact_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_input_market_share_artifact_id ON derived.run_artifacts USING btree (input_market_share_artifact_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_kind_run_id ON derived.run_artifacts USING btree (run_kind, run_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_artifacts_run_key ON derived.run_artifacts USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_run_artifacts_kind_run_artifact ON derived.run_artifacts USING btree (run_kind, run_id, artifact_kind) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_run_jobs_kind_status_created_at ON derived.run_jobs USING btree (run_kind, status, created_at);
CREATE INDEX idx_derived_run_jobs_progress_updated_at ON derived.run_jobs USING btree (progress_updated_at);
CREATE UNIQUE INDEX uq_derived_run_jobs_kind_run_id ON derived.run_jobs USING btree (run_kind, run_id);
CREATE INDEX idx_derived_run_progress_events_kind_run_created_at ON derived.run_progress_events USING btree (run_kind, run_id, created_at);
CREATE INDEX idx_derived_run_progress_events_run_key ON derived.run_progress_events USING btree (run_key);
CREATE INDEX idx_derived_simulated_calcutta_payouts_simulated_calcutta_id ON derived.simulated_calcutta_payouts USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_simulated_calcutta_payouts_simulated_position ON derived.simulated_calcutta_payouts USING btree (simulated_calcutta_id, "position") WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_calcutta_scoring_rules_simulated_calcutta ON derived.simulated_calcutta_scoring_rules USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_simulated_calcutta_scoring_rules_simulated_win_index ON derived.simulated_calcutta_scoring_rules USING btree (simulated_calcutta_id, win_index) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_calcuttas_base_calcutta_id ON derived.simulated_calcuttas USING btree (base_calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_calcuttas_created_at ON derived.simulated_calcuttas USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_calcuttas_highlighted_simulated_entry_id ON derived.simulated_calcuttas USING btree (highlighted_simulated_entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_calcuttas_tournament_id ON derived.simulated_calcuttas USING btree (tournament_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_entries_created_at ON derived.simulated_entries USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_entries_simulated_calcutta_id ON derived.simulated_entries USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_entry_teams_simulated_entry_id ON derived.simulated_entry_teams USING btree (simulated_entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulated_entry_teams_team_id ON derived.simulated_entry_teams USING btree (team_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_simulated_entry_teams_entry_team ON derived.simulated_entry_teams USING btree (simulated_entry_id, team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_analytics_simulated_tournaments_batch_id ON derived.simulated_teams USING btree (simulated_tournament_id);
CREATE INDEX idx_silver_simulated_tournaments_sim_id ON derived.simulated_teams USING btree (tournament_id, sim_id);
CREATE INDEX idx_silver_simulated_tournaments_team_id ON derived.simulated_teams USING btree (team_id);
CREATE INDEX idx_silver_simulated_tournaments_tournament_id ON derived.simulated_teams USING btree (tournament_id);
CREATE UNIQUE INDEX uq_analytics_simulated_tournaments_batch_sim_team ON derived.simulated_teams USING btree (simulated_tournament_id, sim_id, team_id) WHERE ((deleted_at IS NULL) AND (simulated_tournament_id IS NOT NULL));
CREATE UNIQUE INDEX uq_analytics_simulated_tournaments_legacy_sim_team ON derived.simulated_teams USING btree (tournament_id, sim_id, team_id) WHERE ((deleted_at IS NULL) AND (simulated_tournament_id IS NULL));
CREATE INDEX idx_analytics_tournament_simulation_batches_snapshot_id ON derived.simulated_tournaments USING btree (simulation_state_id);
CREATE INDEX idx_analytics_tournament_simulation_batches_tournament_id ON derived.simulated_tournaments USING btree (tournament_id);
CREATE UNIQUE INDEX uq_analytics_tournament_simulation_batches_natural_key ON derived.simulated_tournaments USING btree (tournament_id, simulation_state_id, n_sims, seed, probability_source_key) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_derived_synthetic_calcutta_cohorts_name ON derived.simulation_cohorts USING btree (name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_run_batches_cohort_id ON derived.simulation_run_batches USING btree (cohort_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_run_batches_created_at ON derived.simulation_run_batches USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_calcutta_id ON derived.simulation_runs USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_cohort_id ON derived.simulation_runs USING btree (cohort_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_created_at ON derived.simulation_runs USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_focus_snapshot_entry_id ON derived.simulation_runs USING btree (focus_snapshot_entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_run_key ON derived.simulation_runs USING btree (run_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_derived_simulation_runs_simulated_calcutta_id ON derived.simulation_runs USING btree (simulated_calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_analytics_tournament_state_snapshot_teams_snapshot_id ON derived.simulation_state_teams USING btree (simulation_state_id);
CREATE INDEX idx_analytics_tournament_state_snapshot_teams_team_id ON derived.simulation_state_teams USING btree (team_id);
CREATE INDEX idx_analytics_tournament_state_snapshots_tournament_id ON derived.simulation_states USING btree (tournament_id);

-- Lab indexes
CREATE INDEX idx_lab_entries_calcutta_id ON lab.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_created_at ON lab.entries USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_investment_model_id ON lab.entries USING btree (investment_model_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_entries_state ON lab.entries USING btree (state) WHERE (state <> 'complete'::text);
CREATE UNIQUE INDEX uq_lab_entries_model_calcutta_state ON lab.entries USING btree (investment_model_id, calcutta_id, starting_state_key) WHERE (deleted_at IS NULL);
CREATE INDEX idx_evaluation_entry_results_evaluation_id ON lab.evaluation_entry_results USING btree (evaluation_id);
CREATE INDEX idx_lab_evaluations_created_at ON lab.evaluations USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_evaluations_entry_id ON lab.evaluations USING btree (entry_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_lab_evaluations_entry_sims_seed ON lab.evaluations USING btree (entry_id, n_sims, seed) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_investment_models_created_at ON lab.investment_models USING btree (created_at DESC) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_investment_models_kind ON lab.investment_models USING btree (kind) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_lab_investment_models_name ON lab.investment_models USING btree (name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_lab_pipeline_calcutta_runs_calcutta_id ON lab.pipeline_calcutta_runs USING btree (calcutta_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_pipeline_run_id ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id);
CREATE INDEX idx_lab_pipeline_calcutta_runs_status ON lab.pipeline_calcutta_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));
CREATE UNIQUE INDEX uq_lab_pipeline_calcutta_runs_pipeline_calcutta ON lab.pipeline_calcutta_runs USING btree (pipeline_run_id, calcutta_id);
CREATE INDEX idx_lab_pipeline_runs_created_at ON lab.pipeline_runs USING btree (created_at DESC);
CREATE INDEX idx_lab_pipeline_runs_investment_model_id ON lab.pipeline_runs USING btree (investment_model_id);
CREATE INDEX idx_lab_pipeline_runs_status ON lab.pipeline_runs USING btree (status) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));

-- ============================================================================
-- 14. FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Core auth table FKs (moved from public to core)
ALTER TABLE ONLY core.api_keys ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.auth_sessions ADD CONSTRAINT auth_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);
ALTER TABLE ONLY core.grants ADD CONSTRAINT grants_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_label_id_fkey FOREIGN KEY (label_id) REFERENCES core.labels(id);
ALTER TABLE ONLY core.label_permissions ADD CONSTRAINT label_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES core.permissions(id);

-- Core domain FKs (owner_id and user_id now reference core.users)
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.calcuttas ADD CONSTRAINT calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY core.entries ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.entry_teams ADD CONSTRAINT entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_school_id_fkey FOREIGN KEY (school_id) REFERENCES core.schools(id);
ALTER TABLE ONLY core.teams ADD CONSTRAINT teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_competition_id_fkey FOREIGN KEY (competition_id) REFERENCES core.competitions(id);
ALTER TABLE ONLY core.tournaments ADD CONSTRAINT tournaments_season_id_fkey FOREIGN KEY (season_id) REFERENCES core.seasons(id);
ALTER TABLE ONLY core.team_kenpom_stats ADD CONSTRAINT team_kenpom_stats_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.payouts ADD CONSTRAINT payouts_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_snapshots ADD CONSTRAINT calcutta_snapshots_base_calcutta_id_fkey FOREIGN KEY (base_calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY core.calcutta_snapshot_entries ADD CONSTRAINT calcutta_snapshot_entries_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_snapshot_entries ADD CONSTRAINT calcutta_snapshot_entries_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id);
ALTER TABLE ONLY core.calcutta_snapshot_entry_teams ADD CONSTRAINT calcutta_snapshot_entry_teams_calcutta_snapshot_entry_id_fkey FOREIGN KEY (calcutta_snapshot_entry_id) REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_snapshot_entry_teams ADD CONSTRAINT calcutta_snapshot_entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY core.calcutta_snapshot_payouts ADD CONSTRAINT calcutta_snapshot_payouts_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;
ALTER TABLE ONLY core.calcutta_snapshot_scoring_rules ADD CONSTRAINT calcutta_snapshot_scoring_rules_calcutta_snapshot_id_fkey FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE;

-- Derived FKs
ALTER TABLE ONLY derived.simulation_states ADD CONSTRAINT tournament_state_snapshots_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT tournament_state_snapshot_tea_tournament_state_snapshot_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulation_state_teams ADD CONSTRAINT tournament_state_snapshot_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT tournament_simulation_batches_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.simulated_tournaments ADD CONSTRAINT tournament_simulation_batches_tournament_state_snapshot_id_fkey FOREIGN KEY (simulation_state_id) REFERENCES derived.simulation_states(id);
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_teams_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulated_teams ADD CONSTRAINT simulated_tournaments_tournament_simulation_batch_id_fkey FOREIGN KEY (simulated_tournament_id) REFERENCES derived.simulated_tournaments(id);
ALTER TABLE ONLY derived.game_outcome_runs ADD CONSTRAINT game_outcome_runs_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.game_outcome_runs ADD CONSTRAINT game_outcome_runs_prediction_model_id_fkey FOREIGN KEY (prediction_model_id) REFERENCES derived.prediction_models(id);
ALTER TABLE ONLY derived.calcutta_evaluation_runs ADD CONSTRAINT calcutta_evaluation_runs_tournament_simulation_batch_id_fkey FOREIGN KEY (simulated_tournament_id) REFERENCES derived.simulated_tournaments(id);
ALTER TABLE ONLY derived.calcutta_evaluation_runs ADD CONSTRAINT fk_analytics_calcutta_evaluation_runs_calcutta_snapshot_id FOREIGN KEY (calcutta_snapshot_id) REFERENCES core.calcutta_snapshots(id);
ALTER TABLE ONLY derived.calcutta_evaluation_runs ADD CONSTRAINT fk_derived_calcutta_evaluation_runs_simulated_calcutta_id FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);
ALTER TABLE ONLY derived.entry_performance ADD CONSTRAINT entry_performance_calcutta_evaluation_run_id_fkey FOREIGN KEY (calcutta_evaluation_run_id) REFERENCES derived.calcutta_evaluation_runs(id);
ALTER TABLE ONLY derived.entry_simulation_outcomes ADD CONSTRAINT entry_simulation_outcomes_calcutta_evaluation_run_id_fkey FOREIGN KEY (calcutta_evaluation_run_id) REFERENCES derived.calcutta_evaluation_runs(id);
ALTER TABLE ONLY derived.entry_bids ADD CONSTRAINT entry_bids_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.entry_bids ADD CONSTRAINT entry_bids_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.detailed_investment_report ADD CONSTRAINT detailed_investment_report_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.optimization_runs ADD CONSTRAINT optimization_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_game_outcomes ADD CONSTRAINT predicted_game_outcomes_run_id_fkey FOREIGN KEY (run_id) REFERENCES derived.game_outcome_runs(id);
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.predicted_market_share ADD CONSTRAINT predicted_market_share_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_input_advancement_artifact_id_fkey FOREIGN KEY (input_advancement_artifact_id) REFERENCES derived.run_artifacts(id);
ALTER TABLE ONLY derived.run_artifacts ADD CONSTRAINT run_artifacts_input_market_share_artifact_id_fkey FOREIGN KEY (input_market_share_artifact_id) REFERENCES derived.run_artifacts(id);
ALTER TABLE ONLY derived.simulated_calcuttas ADD CONSTRAINT simulated_calcuttas_base_calcutta_id_fkey FOREIGN KEY (base_calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY derived.simulated_calcuttas ADD CONSTRAINT simulated_calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE ONLY derived.simulated_calcuttas ADD CONSTRAINT simulated_calcuttas_highlighted_simulated_entry_id_fkey FOREIGN KEY (highlighted_simulated_entry_id) REFERENCES derived.simulated_entries(id);
ALTER TABLE ONLY derived.simulated_entries ADD CONSTRAINT simulated_entries_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulated_entries ADD CONSTRAINT simulated_entries_source_entry_id_fkey FOREIGN KEY (source_entry_id) REFERENCES core.entries(id);
ALTER TABLE ONLY derived.simulated_entry_teams ADD CONSTRAINT simulated_entry_teams_simulated_entry_id_fkey FOREIGN KEY (simulated_entry_id) REFERENCES derived.simulated_entries(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulated_entry_teams ADD CONSTRAINT simulated_entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE ONLY derived.simulated_calcutta_payouts ADD CONSTRAINT simulated_calcutta_payouts_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulated_calcutta_scoring_rules ADD CONSTRAINT simulated_calcutta_scoring_rules_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE;
ALTER TABLE ONLY derived.simulation_run_batches ADD CONSTRAINT simulation_run_batches_cohort_id_fkey FOREIGN KEY (cohort_id) REFERENCES derived.simulation_cohorts(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_cohort_id_fkey FOREIGN KEY (cohort_id) REFERENCES derived.simulation_cohorts(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_game_outcome_run_id_fkey FOREIGN KEY (game_outcome_run_id) REFERENCES derived.game_outcome_runs(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_calcutta_evaluation_run_id_fkey FOREIGN KEY (calcutta_evaluation_run_id) REFERENCES derived.calcutta_evaluation_runs(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT simulation_runs_simulation_run_batch_id_fkey FOREIGN KEY (simulation_run_batch_id) REFERENCES derived.simulation_run_batches(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT fk_derived_simulation_runs_focus_snapshot_entry_id FOREIGN KEY (focus_snapshot_entry_id) REFERENCES core.calcutta_snapshot_entries(id);
ALTER TABLE ONLY derived.simulation_runs ADD CONSTRAINT fk_derived_simulation_runs_simulated_calcutta_id FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);
ALTER TABLE ONLY derived.optimized_entries ADD CONSTRAINT optimized_entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY derived.optimized_entries ADD CONSTRAINT optimized_entries_game_outcome_run_id_fkey FOREIGN KEY (game_outcome_run_id) REFERENCES derived.game_outcome_runs(id);

-- Lab FKs
ALTER TABLE ONLY lab.entries ADD CONSTRAINT entries_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);
ALTER TABLE ONLY lab.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY lab.evaluations ADD CONSTRAINT evaluations_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);
ALTER TABLE ONLY lab.evaluations ADD CONSTRAINT evaluations_simulated_calcutta_id_fkey FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);
ALTER TABLE ONLY lab.evaluation_entry_results ADD CONSTRAINT evaluation_entry_results_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);
ALTER TABLE ONLY lab.pipeline_runs ADD CONSTRAINT pipeline_runs_investment_model_id_fkey FOREIGN KEY (investment_model_id) REFERENCES lab.investment_models(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_pipeline_run_id_fkey FOREIGN KEY (pipeline_run_id) REFERENCES lab.pipeline_runs(id) ON DELETE CASCADE;
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);
ALTER TABLE ONLY lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_evaluation_id_fkey FOREIGN KEY (evaluation_id) REFERENCES lab.evaluations(id);

-- ============================================================================
-- 15. TRIGGERS (all standardized to trg_ prefix, all using core.set_updated_at())
-- ============================================================================

-- Core auth table triggers
CREATE TRIGGER trg_core_users_updated_at BEFORE UPDATE ON core.users FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_permissions_updated_at BEFORE UPDATE ON core.permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_labels_updated_at BEFORE UPDATE ON core.labels FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_label_permissions_updated_at BEFORE UPDATE ON core.label_permissions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_api_keys_updated_at BEFORE UPDATE ON core.api_keys FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_auth_sessions_updated_at BEFORE UPDATE ON core.auth_sessions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_bundle_uploads_updated_at BEFORE UPDATE ON core.bundle_uploads FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_grants_updated_at BEFORE UPDATE ON core.grants FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Core domain triggers
CREATE TRIGGER trg_core_seasons_updated_at BEFORE UPDATE ON core.seasons FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_competitions_updated_at BEFORE UPDATE ON core.competitions FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_schools_updated_at BEFORE UPDATE ON core.schools FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_tournaments_updated_at BEFORE UPDATE ON core.tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_teams_updated_at BEFORE UPDATE ON core.teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_team_kenpom_stats_updated_at BEFORE UPDATE ON core.team_kenpom_stats FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcuttas_updated_at BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entries_updated_at BEFORE UPDATE ON core.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entry_teams_updated_at BEFORE UPDATE ON core.entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_payouts_updated_at BEFORE UPDATE ON core.payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_snapshots_updated_at BEFORE UPDATE ON core.calcutta_snapshots FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_snapshot_entries_updated_at BEFORE UPDATE ON core.calcutta_snapshot_entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_snapshot_entry_teams_updated_at BEFORE UPDATE ON core.calcutta_snapshot_entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_snapshot_payouts_updated_at BEFORE UPDATE ON core.calcutta_snapshot_payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_snapshot_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_snapshot_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Derived triggers (updated_at standardized to trg_ prefix and core.set_updated_at())
CREATE TRIGGER trg_derived_simulation_states_updated_at BEFORE UPDATE ON derived.simulation_states FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_state_teams_updated_at BEFORE UPDATE ON derived.simulation_state_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_tournaments_updated_at BEFORE UPDATE ON derived.simulated_tournaments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_teams_updated_at BEFORE UPDATE ON derived.simulated_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_prediction_models_updated_at BEFORE UPDATE ON derived.prediction_models FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_game_outcome_runs_updated_at BEFORE UPDATE ON derived.game_outcome_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_calcutta_evaluation_runs_updated_at BEFORE UPDATE ON derived.calcutta_evaluation_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_entry_performance_updated_at BEFORE UPDATE ON derived.entry_performance FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_entry_simulation_outcomes_updated_at BEFORE UPDATE ON derived.entry_simulation_outcomes FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_entry_bids_updated_at BEFORE UPDATE ON derived.entry_bids FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_detailed_investment_report_updated_at BEFORE UPDATE ON derived.detailed_investment_report FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_optimization_runs_updated_at BEFORE UPDATE ON derived.optimization_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_predicted_game_outcomes_updated_at BEFORE UPDATE ON derived.predicted_game_outcomes FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_predicted_market_share_updated_at BEFORE UPDATE ON derived.predicted_market_share FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_run_artifacts_updated_at BEFORE UPDATE ON derived.run_artifacts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_run_jobs_updated_at BEFORE UPDATE ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_calcuttas_updated_at BEFORE UPDATE ON derived.simulated_calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_entries_updated_at BEFORE UPDATE ON derived.simulated_entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_entry_teams_updated_at BEFORE UPDATE ON derived.simulated_entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_calcutta_payouts_updated_at BEFORE UPDATE ON derived.simulated_calcutta_payouts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulated_calcutta_scoring_rules_updated_at BEFORE UPDATE ON derived.simulated_calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_cohorts_updated_at BEFORE UPDATE ON derived.simulation_cohorts FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_run_batches_updated_at BEFORE UPDATE ON derived.simulation_run_batches FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_simulation_runs_updated_at BEFORE UPDATE ON derived.simulation_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_derived_optimized_entries_updated_at BEFORE UPDATE ON derived.optimized_entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Derived enqueue triggers (unchanged from dump)
CREATE TRIGGER trg_derived_calcutta_evaluation_runs_enqueue_run_job AFTER INSERT ON derived.calcutta_evaluation_runs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_calcutta_evaluation_run();
CREATE TRIGGER trg_derived_game_outcome_runs_enqueue_run_job AFTER INSERT ON derived.game_outcome_runs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_game_outcome_run();
CREATE TRIGGER trg_derived_simulation_runs_enqueue_run_job AFTER INSERT ON derived.simulation_runs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_job_for_simulation_run();
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_insert AFTER INSERT ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_update AFTER UPDATE ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();

-- Lab triggers (standardized to trg_ prefix and core.set_updated_at())
CREATE TRIGGER trg_lab_entries_updated_at BEFORE UPDATE ON lab.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_evaluations_updated_at BEFORE UPDATE ON lab.evaluations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_investment_models_updated_at BEFORE UPDATE ON lab.investment_models FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_pipeline_calcutta_runs_updated_at BEFORE UPDATE ON lab.pipeline_calcutta_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_lab_pipeline_runs_updated_at BEFORE UPDATE ON lab.pipeline_runs FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
