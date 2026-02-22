-- =============================================================================
-- Migration: 20260222000004_database_audit_fixes
-- Fixes: 1 FAIL + 6 actionable PARTIALs from database audit
-- =============================================================================

-- =============================================================================
-- Fix 1: FAIL — Rename eliminated → is_eliminated (3 tables + 2 views)
-- =============================================================================

-- Drop dependent views first (portfolios depends on portfolio_teams)
DROP VIEW IF EXISTS derived.portfolios;
DROP VIEW IF EXISTS derived.portfolio_teams;

-- Rename columns
ALTER TABLE core.teams RENAME COLUMN eliminated TO is_eliminated;
ALTER TABLE derived.simulation_state_teams RENAME COLUMN eliminated TO is_eliminated;
ALTER TABLE derived.simulated_teams RENAME COLUMN eliminated TO is_eliminated;

-- Recreate derived.portfolio_teams with is_eliminated
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
            tt.is_eliminated,
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
                    WHEN (eb.is_eliminated = true) THEN (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision
                    ELSE (core.calcutta_points_for_progress(eb.calcutta_id, eb.tournament_rounds, 0))::double precision
                END) AS expected_points,
            eb.school_id,
            eb.tournament_id,
            eb.seed,
            eb.region,
            eb.byes,
            eb.wins,
            eb.is_eliminated,
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

-- Recreate derived.portfolios (depends on portfolio_teams)
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
-- Fix 2: PARTIAL — Rename 5 auto-generated CHECK constraints
-- =============================================================================

ALTER TABLE core.users RENAME CONSTRAINT users_status_check TO ck_core_users_status;
ALTER TABLE core.bundle_uploads RENAME CONSTRAINT bundle_uploads_status_check TO ck_core_bundle_uploads_status;
ALTER TABLE core.grants RENAME CONSTRAINT grants_scope_id_check TO ck_core_grants_scope_id;
ALTER TABLE core.grants RENAME CONSTRAINT grants_scope_type_check TO ck_core_grants_scope_type;
ALTER TABLE core.grants RENAME CONSTRAINT grants_subject_check TO ck_core_grants_subject;

-- =============================================================================
-- Fix 3: PARTIAL — Add updated_at + deleted_at to core.idempotency_keys
-- =============================================================================

ALTER TABLE core.idempotency_keys
  ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT now() NOT NULL,
  ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE TRIGGER trg_idempotency_keys_set_updated_at
  BEFORE UPDATE ON core.idempotency_keys
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- Fix 4: PARTIAL — Add 3 indexes on lab.pipeline_calcutta_runs FK columns
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_lab_pipeline_calcutta_runs_predictions_job
  ON lab.pipeline_calcutta_runs (predictions_job_id) WHERE predictions_job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_lab_pipeline_calcutta_runs_optimization_job
  ON lab.pipeline_calcutta_runs (optimization_job_id) WHERE optimization_job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_lab_pipeline_calcutta_runs_evaluation_job
  ON lab.pipeline_calcutta_runs (evaluation_job_id) WHERE evaluation_job_id IS NOT NULL;

-- =============================================================================
-- Fix 5: PARTIAL — Remove prevent_calcutta_soft_delete trigger
-- =============================================================================

DROP TRIGGER IF EXISTS trg_calcuttas_prevent_soft_delete ON core.calcuttas;
DROP FUNCTION IF EXISTS core.prevent_calcutta_soft_delete();
