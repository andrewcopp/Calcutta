-- =============================================================================
-- Rollback: 20260222000004_database_audit_fixes
-- =============================================================================

-- =============================================================================
-- Undo Fix 5: Recreate prevent_calcutta_soft_delete trigger
-- =============================================================================

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

CREATE TRIGGER trg_calcuttas_prevent_soft_delete
  BEFORE UPDATE ON core.calcuttas
  FOR EACH ROW EXECUTE FUNCTION core.prevent_calcutta_soft_delete();

-- =============================================================================
-- Undo Fix 4: Drop 3 indexes on lab.pipeline_calcutta_runs FK columns
-- =============================================================================

DROP INDEX IF EXISTS lab.idx_lab_pipeline_calcutta_runs_predictions_job;
DROP INDEX IF EXISTS lab.idx_lab_pipeline_calcutta_runs_optimization_job;
DROP INDEX IF EXISTS lab.idx_lab_pipeline_calcutta_runs_evaluation_job;

-- =============================================================================
-- Undo Fix 3: Remove updated_at + deleted_at from core.idempotency_keys
-- =============================================================================

DROP TRIGGER IF EXISTS trg_idempotency_keys_set_updated_at ON core.idempotency_keys;
ALTER TABLE core.idempotency_keys DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE core.idempotency_keys DROP COLUMN IF EXISTS updated_at;

-- =============================================================================
-- Undo Fix 2: Rename CHECK constraints back to auto-generated names
-- =============================================================================

ALTER TABLE core.users RENAME CONSTRAINT ck_core_users_status TO users_status_check;
ALTER TABLE core.bundle_uploads RENAME CONSTRAINT ck_core_bundle_uploads_status TO bundle_uploads_status_check;
ALTER TABLE core.grants RENAME CONSTRAINT ck_core_grants_scope_id TO grants_scope_id_check;
ALTER TABLE core.grants RENAME CONSTRAINT ck_core_grants_scope_type TO grants_scope_type_check;
ALTER TABLE core.grants RENAME CONSTRAINT ck_core_grants_subject TO grants_subject_check;

-- =============================================================================
-- Undo Fix 1: Rename is_eliminated → eliminated (3 tables + 2 views)
-- =============================================================================

DROP VIEW IF EXISTS derived.portfolios;
DROP VIEW IF EXISTS derived.portfolio_teams;

ALTER TABLE core.teams RENAME COLUMN is_eliminated TO eliminated;
ALTER TABLE derived.simulation_state_teams RENAME COLUMN is_eliminated TO eliminated;
ALTER TABLE derived.simulated_teams RENAME COLUMN is_eliminated TO eliminated;

-- Recreate derived.portfolio_teams with eliminated
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

-- Recreate derived.portfolios
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
