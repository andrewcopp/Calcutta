-- Migration: rename_calcutta_vocab
-- Created: 2026-02-25 23:55:16 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '60s';

-- =============================================================================
-- STEP 1: Drop derived views (they reference old table/column names)
-- =============================================================================

DROP VIEW IF EXISTS derived.portfolios;
DROP VIEW IF EXISTS derived.portfolio_teams;

-- =============================================================================
-- STEP 2: Drop foreign keys that reference tables being renamed
-- (needed before renaming tables to avoid conflicts)
-- =============================================================================

-- FKs referencing core.calcuttas
ALTER TABLE core.calcutta_invitations DROP CONSTRAINT IF EXISTS calcutta_invitations_calcutta_id_fkey;
ALTER TABLE core.calcutta_scoring_rules DROP CONSTRAINT IF EXISTS calcutta_scoring_rules_calcutta_id_fkey;
ALTER TABLE core.entries DROP CONSTRAINT IF EXISTS entries_calcutta_id_fkey;
ALTER TABLE core.payouts DROP CONSTRAINT IF EXISTS payouts_calcutta_id_fkey;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS calcuttas_created_by_fkey;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS calcuttas_owner_id_fkey;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS calcuttas_tournament_id_fkey;

-- FKs referencing core.entries
ALTER TABLE core.entry_teams DROP CONSTRAINT IF EXISTS entry_teams_entry_id_fkey;

-- FKs from lab referencing core.calcuttas
ALTER TABLE lab.entries DROP CONSTRAINT IF EXISTS entries_calcutta_id_fkey;
ALTER TABLE lab.pipeline_calcutta_runs DROP CONSTRAINT IF EXISTS pipeline_calcutta_runs_calcutta_id_fkey;
ALTER TABLE lab.pipeline_calcutta_runs DROP CONSTRAINT IF EXISTS pipeline_calcutta_runs_entry_id_fkey;

-- =============================================================================
-- STEP 3: Drop triggers on tables being renamed
-- =============================================================================

DROP TRIGGER IF EXISTS trg_calcuttas_immutable_created_by ON core.calcuttas;
DROP TRIGGER IF EXISTS trg_core_calcuttas_updated_at ON core.calcuttas;
DROP TRIGGER IF EXISTS trg_core_calcutta_invitations_updated_at ON core.calcutta_invitations;
DROP TRIGGER IF EXISTS trg_core_calcutta_scoring_rules_updated_at ON core.calcutta_scoring_rules;
DROP TRIGGER IF EXISTS trg_core_entries_updated_at ON core.entries;
DROP TRIGGER IF EXISTS trg_core_entry_teams_updated_at ON core.entry_teams;

-- =============================================================================
-- STEP 4: Drop indexes on tables being renamed
-- =============================================================================

-- core.calcuttas indexes
DROP INDEX IF EXISTS core.idx_calcuttas_active;
DROP INDEX IF EXISTS core.idx_calcuttas_created_by;
DROP INDEX IF EXISTS core.idx_core_calcuttas_budget_points;
DROP INDEX IF EXISTS core.idx_core_calcuttas_owner_id;
DROP INDEX IF EXISTS core.idx_core_calcuttas_tournament_id;

-- core.calcutta_invitations indexes
DROP INDEX IF EXISTS core.idx_calcutta_invitations_active;
DROP INDEX IF EXISTS core.idx_calcutta_invitations_calcutta_id_active;
DROP INDEX IF EXISTS core.idx_calcutta_invitations_invited_by;
DROP INDEX IF EXISTS core.idx_calcutta_invitations_user_id;
DROP INDEX IF EXISTS core.uq_calcutta_invitations_calcutta_user;

-- core.calcutta_scoring_rules indexes
DROP INDEX IF EXISTS core.idx_calcutta_scoring_rules_active;
DROP INDEX IF EXISTS core.idx_core_calcutta_scoring_rules_calcutta_id;

-- core.entries indexes
DROP INDEX IF EXISTS core.idx_entries_active;
DROP INDEX IF EXISTS core.idx_entries_calcutta_id_active;
DROP INDEX IF EXISTS core.idx_core_entries_user_id;
DROP INDEX IF EXISTS core.uq_entries_user_calcutta;

-- core.entry_teams indexes
DROP INDEX IF EXISTS core.idx_entry_teams_active;
DROP INDEX IF EXISTS core.idx_entry_teams_entry_id_active;
DROP INDEX IF EXISTS core.idx_entry_teams_team_id_active;
DROP INDEX IF EXISTS core.uq_core_entry_teams_entry_team;

-- core.payouts indexes
DROP INDEX IF EXISTS core.idx_payouts_active;
DROP INDEX IF EXISTS core.idx_core_payouts_calcutta_id;

-- =============================================================================
-- STEP 5: Drop check constraints on tables being renamed
-- =============================================================================

ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_budget_positive;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_max_bid_points_le_budget;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_max_bid_points_positive;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_max_teams_gte_min;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_min_teams;
ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS ck_core_calcuttas_visibility;
ALTER TABLE core.calcutta_invitations DROP CONSTRAINT IF EXISTS ck_calcutta_invitations_status;
ALTER TABLE core.calcutta_scoring_rules DROP CONSTRAINT IF EXISTS chk_scoring_rules_points_nonneg;
ALTER TABLE core.calcutta_scoring_rules DROP CONSTRAINT IF EXISTS chk_scoring_rules_win_index_nonneg;
ALTER TABLE core.entry_teams DROP CONSTRAINT IF EXISTS chk_entry_teams_bid_positive;

-- =============================================================================
-- STEP 6: Drop unique constraints on tables being renamed
-- =============================================================================

ALTER TABLE core.calcutta_scoring_rules DROP CONSTRAINT IF EXISTS uq_core_calcutta_scoring_rules;
ALTER TABLE core.payouts DROP CONSTRAINT IF EXISTS uq_core_payouts_calcutta_position;

-- =============================================================================
-- STEP 7: Drop primary keys on tables being renamed
-- =============================================================================

ALTER TABLE core.calcuttas DROP CONSTRAINT IF EXISTS calcuttas_pkey;
ALTER TABLE core.calcutta_invitations DROP CONSTRAINT IF EXISTS calcutta_invitations_pkey;
ALTER TABLE core.calcutta_scoring_rules DROP CONSTRAINT IF EXISTS calcutta_scoring_rules_pkey;
ALTER TABLE core.entries DROP CONSTRAINT IF EXISTS entries_pkey;
ALTER TABLE core.entry_teams DROP CONSTRAINT IF EXISTS entry_teams_pkey;

-- =============================================================================
-- STEP 8: Rename tables
-- =============================================================================

ALTER TABLE core.calcuttas RENAME TO pools;
ALTER TABLE core.entries RENAME TO portfolios;
ALTER TABLE core.entry_teams RENAME TO investments;
ALTER TABLE core.calcutta_scoring_rules RENAME TO pool_scoring_rules;
ALTER TABLE core.calcutta_invitations RENAME TO pool_invitations;

-- =============================================================================
-- STEP 9: Rename columns
-- =============================================================================

-- core.pools (was core.calcuttas)
ALTER TABLE core.pools RENAME COLUMN max_bid_points TO max_investment_credits;
ALTER TABLE core.pools RENAME COLUMN budget_points TO budget_credits;

-- core.portfolios (was core.entries)
ALTER TABLE core.portfolios RENAME COLUMN calcutta_id TO pool_id;

-- core.investments (was core.entry_teams)
ALTER TABLE core.investments RENAME COLUMN entry_id TO portfolio_id;
ALTER TABLE core.investments RENAME COLUMN bid_points TO credits;

-- core.pool_scoring_rules (was core.calcutta_scoring_rules)
ALTER TABLE core.pool_scoring_rules RENAME COLUMN calcutta_id TO pool_id;

-- core.pool_invitations (was core.calcutta_invitations)
ALTER TABLE core.pool_invitations RENAME COLUMN calcutta_id TO pool_id;

-- core.payouts
ALTER TABLE core.payouts RENAME COLUMN calcutta_id TO pool_id;

-- =============================================================================
-- STEP 10: Recreate primary keys with new names
-- =============================================================================

ALTER TABLE core.pools ADD CONSTRAINT pools_pkey PRIMARY KEY (id);
ALTER TABLE core.pool_invitations ADD CONSTRAINT pool_invitations_pkey PRIMARY KEY (id);
ALTER TABLE core.pool_scoring_rules ADD CONSTRAINT pool_scoring_rules_pkey PRIMARY KEY (id);
ALTER TABLE core.portfolios ADD CONSTRAINT portfolios_pkey PRIMARY KEY (id);
ALTER TABLE core.investments ADD CONSTRAINT investments_pkey PRIMARY KEY (id);

-- =============================================================================
-- STEP 11: Recreate unique constraints with new names
-- =============================================================================

ALTER TABLE core.pool_scoring_rules ADD CONSTRAINT uq_core_pool_scoring_rules UNIQUE (pool_id, win_index);
ALTER TABLE core.payouts ADD CONSTRAINT uq_core_payouts_pool_position UNIQUE (pool_id, "position");

-- =============================================================================
-- STEP 12: Recreate check constraints with new names
-- =============================================================================

ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_budget_positive CHECK (budget_credits > 0);
ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_max_investment_credits_le_budget CHECK (max_investment_credits <= budget_credits);
ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_max_investment_credits_positive CHECK (max_investment_credits > 0);
ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_max_teams_gte_min CHECK (max_teams >= min_teams);
ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_min_teams CHECK (min_teams >= 1);
ALTER TABLE core.pools ADD CONSTRAINT ck_core_pools_visibility CHECK (visibility = ANY (ARRAY['public'::text, 'unlisted'::text, 'private'::text]));
ALTER TABLE core.pool_invitations ADD CONSTRAINT ck_pool_invitations_status CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'revoked'::text]));
ALTER TABLE core.pool_scoring_rules ADD CONSTRAINT chk_scoring_rules_points_nonneg CHECK (points_awarded >= 0);
ALTER TABLE core.pool_scoring_rules ADD CONSTRAINT chk_scoring_rules_win_index_nonneg CHECK (win_index >= 0);
ALTER TABLE core.investments ADD CONSTRAINT chk_investments_credits_positive CHECK (credits > 0);

-- =============================================================================
-- STEP 13: Recreate indexes with new names
-- =============================================================================

-- core.pools indexes
CREATE INDEX idx_pools_active ON core.pools USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_pools_created_by ON core.pools USING btree (created_by);
CREATE INDEX idx_core_pools_budget_credits ON core.pools USING btree (budget_credits);
CREATE INDEX idx_core_pools_owner_id ON core.pools USING btree (owner_id);
CREATE INDEX idx_core_pools_tournament_id ON core.pools USING btree (tournament_id);

-- core.pool_invitations indexes
CREATE INDEX idx_pool_invitations_active ON core.pool_invitations USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_pool_invitations_pool_id_active ON core.pool_invitations USING btree (pool_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_pool_invitations_invited_by ON core.pool_invitations USING btree (invited_by);
CREATE INDEX idx_pool_invitations_user_id ON core.pool_invitations USING btree (user_id);
CREATE UNIQUE INDEX uq_pool_invitations_pool_user ON core.pool_invitations USING btree (pool_id, user_id) WHERE (deleted_at IS NULL);

-- core.pool_scoring_rules indexes
CREATE INDEX idx_pool_scoring_rules_active ON core.pool_scoring_rules USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_pool_scoring_rules_pool_id ON core.pool_scoring_rules USING btree (pool_id);

-- core.portfolios indexes
CREATE INDEX idx_portfolios_active ON core.portfolios USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_portfolios_pool_id_active ON core.portfolios USING btree (pool_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_portfolios_user_id ON core.portfolios USING btree (user_id);
CREATE UNIQUE INDEX uq_portfolios_user_pool ON core.portfolios USING btree (user_id, pool_id) WHERE ((user_id IS NOT NULL) AND (deleted_at IS NULL));

-- core.investments indexes
CREATE INDEX idx_investments_active ON core.investments USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_investments_portfolio_id_active ON core.investments USING btree (portfolio_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_investments_team_id_active ON core.investments USING btree (team_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_core_investments_portfolio_team ON core.investments USING btree (portfolio_id, team_id) WHERE (deleted_at IS NULL);

-- core.payouts indexes
CREATE INDEX idx_payouts_active ON core.payouts USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_payouts_pool_id ON core.payouts USING btree (pool_id);

-- =============================================================================
-- STEP 14: Recreate foreign keys with new names
-- =============================================================================

-- core.pools FKs
ALTER TABLE core.pools ADD CONSTRAINT pools_created_by_fkey FOREIGN KEY (created_by) REFERENCES core.users(id);
ALTER TABLE core.pools ADD CONSTRAINT pools_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES core.users(id);
ALTER TABLE core.pools ADD CONSTRAINT pools_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);

-- core.pool_invitations FKs
ALTER TABLE core.pool_invitations ADD CONSTRAINT pool_invitations_pool_id_fkey FOREIGN KEY (pool_id) REFERENCES core.pools(id);
ALTER TABLE core.pool_invitations ADD CONSTRAINT pool_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES core.users(id);
ALTER TABLE core.pool_invitations ADD CONSTRAINT pool_invitations_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);

-- core.pool_scoring_rules FKs
ALTER TABLE core.pool_scoring_rules ADD CONSTRAINT pool_scoring_rules_pool_id_fkey FOREIGN KEY (pool_id) REFERENCES core.pools(id) ON DELETE CASCADE;

-- core.portfolios FKs
ALTER TABLE core.portfolios ADD CONSTRAINT portfolios_pool_id_fkey FOREIGN KEY (pool_id) REFERENCES core.pools(id);
ALTER TABLE core.portfolios ADD CONSTRAINT portfolios_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);

-- core.investments FKs
ALTER TABLE core.investments ADD CONSTRAINT investments_portfolio_id_fkey FOREIGN KEY (portfolio_id) REFERENCES core.portfolios(id) ON DELETE CASCADE;
ALTER TABLE core.investments ADD CONSTRAINT investments_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);

-- core.payouts FKs
ALTER TABLE core.payouts ADD CONSTRAINT payouts_pool_id_fkey FOREIGN KEY (pool_id) REFERENCES core.pools(id) ON DELETE CASCADE;

-- lab FKs referencing core.pools (was core.calcuttas)
ALTER TABLE lab.entries ADD CONSTRAINT entries_pool_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.pools(id);
ALTER TABLE lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_pool_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.pools(id);
ALTER TABLE lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);

-- =============================================================================
-- STEP 15: Recreate triggers with new names
-- =============================================================================

CREATE TRIGGER trg_pools_immutable_created_by BEFORE UPDATE ON core.pools FOR EACH ROW EXECUTE FUNCTION core.immutable_created_by();
CREATE TRIGGER trg_core_pools_updated_at BEFORE UPDATE ON core.pools FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_pool_invitations_updated_at BEFORE UPDATE ON core.pool_invitations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_pool_scoring_rules_updated_at BEFORE UPDATE ON core.pool_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_portfolios_updated_at BEFORE UPDATE ON core.portfolios FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_investments_updated_at BEFORE UPDATE ON core.investments FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- STEP 16: Rename function
-- =============================================================================

DROP FUNCTION IF EXISTS core.calcutta_points_for_progress(uuid, integer, integer);

CREATE FUNCTION core.pool_returns_for_progress(p_pool_id uuid, p_wins integer, p_byes integer DEFAULT 0) RETURNS integer
    LANGUAGE sql STABLE
    AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.pool_scoring_rules r
    WHERE r.pool_id = p_pool_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;

-- =============================================================================
-- STEP 17: Recreate derived views with new names
-- =============================================================================

CREATE VIEW derived.ownership_details AS
 WITH portfolio_investments AS (
         SELECT p.id AS portfolio_id,
            p.pool_id,
            inv.team_id,
            (inv.credits)::double precision AS credits,
            inv.created_at AS investment_created_at,
            inv.updated_at AS investment_updated_at,
            sum((inv.credits)::double precision) OVER (PARTITION BY p.pool_id, inv.team_id) AS team_total_credits,
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
            GREATEST(p.updated_at, inv.updated_at, tt.updated_at) AS derived_updated_at
           FROM ((((core.portfolios p
             JOIN core.investments inv ON (((inv.portfolio_id = p.id) AND (inv.deleted_at IS NULL))))
             JOIN core.teams tt ON (((tt.id = inv.team_id) AND (tt.deleted_at IS NULL))))
             JOIN core.tournaments t ON (((t.id = tt.tournament_id) AND (t.deleted_at IS NULL))))
             LEFT JOIN core.schools s ON (((s.id = tt.school_id) AND (s.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
        ), investment_returns AS (
         SELECT pi.portfolio_id,
            pi.pool_id,
            pi.team_id,
            pi.credits,
            pi.team_total_credits,
                CASE
                    WHEN (pi.team_total_credits > (0)::double precision) THEN (pi.credits / pi.team_total_credits)
                    ELSE (0)::double precision
                END AS ownership_percentage,
            (
                CASE
                    WHEN (pi.team_total_credits > (0)::double precision) THEN (pi.credits / pi.team_total_credits)
                    ELSE (0)::double precision
                END * (core.pool_returns_for_progress(pi.pool_id, pi.wins, pi.byes))::double precision) AS actual_returns,
            (
                CASE
                    WHEN (pi.team_total_credits > (0)::double precision) THEN (pi.credits / pi.team_total_credits)
                    ELSE (0)::double precision
                END *
                CASE
                    WHEN (pi.is_eliminated = true) THEN (core.pool_returns_for_progress(pi.pool_id, pi.wins, pi.byes))::double precision
                    ELSE (core.pool_returns_for_progress(pi.pool_id, pi.tournament_rounds, 0))::double precision
                END) AS expected_returns,
            pi.school_id,
            pi.tournament_id,
            pi.seed,
            pi.region,
            pi.byes,
            pi.wins,
            pi.is_eliminated,
            pi.team_created_at,
            pi.team_updated_at,
            pi.school_name,
            pi.investment_created_at AS created_at,
            pi.derived_updated_at AS updated_at,
            NULL::timestamp with time zone AS deleted_at
           FROM portfolio_investments pi
        )
 SELECT concat(portfolio_id, '-', team_id) AS id,
    portfolio_id,
    team_id,
    ownership_percentage,
    actual_returns,
    expected_returns,
    created_at,
    updated_at,
    deleted_at
   FROM investment_returns;

CREATE VIEW derived.ownership_summaries AS
 WITH portfolio_totals AS (
         SELECT od.portfolio_id,
            sum(od.expected_returns) AS maximum_returns,
            max(od.updated_at) AS updated_at
           FROM derived.ownership_details od
          GROUP BY od.portfolio_id
        )
 SELECT p.id,
    p.id AS portfolio_id,
    COALESCE(pt.maximum_returns, (0)::double precision) AS maximum_returns,
    p.created_at,
    COALESCE(pt.updated_at, p.updated_at) AS updated_at,
    NULL::timestamp with time zone AS deleted_at
   FROM (core.portfolios p
     LEFT JOIN portfolio_totals pt ON ((pt.portfolio_id = p.id)))
  WHERE (p.deleted_at IS NULL);

-- =============================================================================
-- STEP 18: Update grants scope_type
-- =============================================================================

-- Update existing grants data
UPDATE core.grants SET scope_type = 'pool' WHERE scope_type = 'calcutta';

-- Drop and recreate the CHECK constraint with new value
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS ck_core_grants_scope_type;
ALTER TABLE core.grants ADD CONSTRAINT ck_core_grants_scope_type CHECK (scope_type = ANY (ARRAY['global'::text, 'pool'::text, 'tournament'::text]));
