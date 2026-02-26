-- Rollback: rename_calcutta_vocab
-- Created: 2026-02-25 23:55:16 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '60s';

-- =============================================================================
-- STEP 1: Drop derived views (they reference new names)
-- =============================================================================

DROP VIEW IF EXISTS derived.ownership_details;
DROP VIEW IF EXISTS derived.ownership_summaries;

-- =============================================================================
-- STEP 2: Revert grants scope_type
-- =============================================================================

UPDATE core.grants SET scope_type = 'calcutta' WHERE scope_type = 'pool';
ALTER TABLE core.grants DROP CONSTRAINT IF EXISTS ck_core_grants_scope_type;
ALTER TABLE core.grants ADD CONSTRAINT ck_core_grants_scope_type CHECK (scope_type = ANY (ARRAY['global'::text, 'calcutta'::text, 'tournament'::text]));

-- =============================================================================
-- STEP 3: Drop renamed function, recreate original
-- =============================================================================

DROP FUNCTION IF EXISTS core.pool_returns_for_progress(uuid, integer, integer);

CREATE FUNCTION core.calcutta_points_for_progress(p_calcutta_id uuid, p_wins integer, p_byes integer DEFAULT 0) RETURNS integer
    LANGUAGE sql STABLE
    AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.calcutta_scoring_rules r
    WHERE r.calcutta_id = p_calcutta_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;

-- =============================================================================
-- STEP 4: Drop triggers
-- =============================================================================

DROP TRIGGER IF EXISTS trg_pools_immutable_created_by ON core.pools;
DROP TRIGGER IF EXISTS trg_core_pools_updated_at ON core.pools;
DROP TRIGGER IF EXISTS trg_core_pool_invitations_updated_at ON core.pool_invitations;
DROP TRIGGER IF EXISTS trg_core_pool_scoring_rules_updated_at ON core.pool_scoring_rules;
DROP TRIGGER IF EXISTS trg_core_portfolios_updated_at ON core.portfolios;
DROP TRIGGER IF EXISTS trg_core_investments_updated_at ON core.investments;

-- =============================================================================
-- STEP 5: Drop foreign keys
-- =============================================================================

ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS pools_created_by_fkey;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS pools_owner_id_fkey;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS pools_tournament_id_fkey;
ALTER TABLE core.pool_invitations DROP CONSTRAINT IF EXISTS pool_invitations_pool_id_fkey;
ALTER TABLE core.pool_invitations DROP CONSTRAINT IF EXISTS pool_invitations_invited_by_fkey;
ALTER TABLE core.pool_invitations DROP CONSTRAINT IF EXISTS pool_invitations_user_id_fkey;
ALTER TABLE core.pool_scoring_rules DROP CONSTRAINT IF EXISTS pool_scoring_rules_pool_id_fkey;
ALTER TABLE core.portfolios DROP CONSTRAINT IF EXISTS portfolios_pool_id_fkey;
ALTER TABLE core.portfolios DROP CONSTRAINT IF EXISTS portfolios_user_id_fkey;
ALTER TABLE core.investments DROP CONSTRAINT IF EXISTS investments_portfolio_id_fkey;
ALTER TABLE core.investments DROP CONSTRAINT IF EXISTS investments_team_id_fkey;
ALTER TABLE core.payouts DROP CONSTRAINT IF EXISTS payouts_pool_id_fkey;
ALTER TABLE lab.entries DROP CONSTRAINT IF EXISTS entries_pool_id_fkey;
ALTER TABLE lab.pipeline_calcutta_runs DROP CONSTRAINT IF EXISTS pipeline_calcutta_runs_pool_id_fkey;
ALTER TABLE lab.pipeline_calcutta_runs DROP CONSTRAINT IF EXISTS pipeline_calcutta_runs_entry_id_fkey;

-- =============================================================================
-- STEP 6: Drop indexes
-- =============================================================================

DROP INDEX IF EXISTS core.idx_pools_active;
DROP INDEX IF EXISTS core.idx_pools_created_by;
DROP INDEX IF EXISTS core.idx_core_pools_budget_credits;
DROP INDEX IF EXISTS core.idx_core_pools_owner_id;
DROP INDEX IF EXISTS core.idx_core_pools_tournament_id;
DROP INDEX IF EXISTS core.idx_pool_invitations_active;
DROP INDEX IF EXISTS core.idx_pool_invitations_pool_id_active;
DROP INDEX IF EXISTS core.idx_pool_invitations_invited_by;
DROP INDEX IF EXISTS core.idx_pool_invitations_user_id;
DROP INDEX IF EXISTS core.uq_pool_invitations_pool_user;
DROP INDEX IF EXISTS core.idx_pool_scoring_rules_active;
DROP INDEX IF EXISTS core.idx_core_pool_scoring_rules_pool_id;
DROP INDEX IF EXISTS core.idx_portfolios_active;
DROP INDEX IF EXISTS core.idx_portfolios_pool_id_active;
DROP INDEX IF EXISTS core.idx_core_portfolios_user_id;
DROP INDEX IF EXISTS core.uq_portfolios_user_pool;
DROP INDEX IF EXISTS core.idx_investments_active;
DROP INDEX IF EXISTS core.idx_investments_portfolio_id_active;
DROP INDEX IF EXISTS core.idx_investments_team_id_active;
DROP INDEX IF EXISTS core.uq_core_investments_portfolio_team;
DROP INDEX IF EXISTS core.idx_payouts_active;
DROP INDEX IF EXISTS core.idx_core_payouts_pool_id;

-- =============================================================================
-- STEP 7: Drop check constraints
-- =============================================================================

ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_budget_positive;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_max_investment_credits_le_budget;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_max_investment_credits_positive;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_max_teams_gte_min;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_min_teams;
ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS ck_core_pools_visibility;
ALTER TABLE core.pool_invitations DROP CONSTRAINT IF EXISTS ck_pool_invitations_status;
ALTER TABLE core.pool_scoring_rules DROP CONSTRAINT IF EXISTS chk_scoring_rules_points_nonneg;
ALTER TABLE core.pool_scoring_rules DROP CONSTRAINT IF EXISTS chk_scoring_rules_win_index_nonneg;
ALTER TABLE core.investments DROP CONSTRAINT IF EXISTS chk_investments_credits_positive;

-- =============================================================================
-- STEP 8: Drop unique constraints
-- =============================================================================

ALTER TABLE core.pool_scoring_rules DROP CONSTRAINT IF EXISTS uq_core_pool_scoring_rules;
ALTER TABLE core.payouts DROP CONSTRAINT IF EXISTS uq_core_payouts_pool_position;

-- =============================================================================
-- STEP 9: Drop primary keys
-- =============================================================================

ALTER TABLE core.pools DROP CONSTRAINT IF EXISTS pools_pkey;
ALTER TABLE core.pool_invitations DROP CONSTRAINT IF EXISTS pool_invitations_pkey;
ALTER TABLE core.pool_scoring_rules DROP CONSTRAINT IF EXISTS pool_scoring_rules_pkey;
ALTER TABLE core.portfolios DROP CONSTRAINT IF EXISTS portfolios_pkey;
ALTER TABLE core.investments DROP CONSTRAINT IF EXISTS investments_pkey;

-- =============================================================================
-- STEP 10: Rename columns back
-- =============================================================================

ALTER TABLE core.pools RENAME COLUMN max_investment_credits TO max_bid_points;
ALTER TABLE core.pools RENAME COLUMN budget_credits TO budget_points;
ALTER TABLE core.portfolios RENAME COLUMN pool_id TO calcutta_id;
ALTER TABLE core.investments RENAME COLUMN portfolio_id TO entry_id;
ALTER TABLE core.investments RENAME COLUMN credits TO bid_points;
ALTER TABLE core.pool_scoring_rules RENAME COLUMN pool_id TO calcutta_id;
ALTER TABLE core.pool_invitations RENAME COLUMN pool_id TO calcutta_id;
ALTER TABLE core.payouts RENAME COLUMN pool_id TO calcutta_id;

-- =============================================================================
-- STEP 11: Rename tables back
-- =============================================================================

ALTER TABLE core.pools RENAME TO calcuttas;
ALTER TABLE core.portfolios RENAME TO entries;
ALTER TABLE core.investments RENAME TO entry_teams;
ALTER TABLE core.pool_scoring_rules RENAME TO calcutta_scoring_rules;
ALTER TABLE core.pool_invitations RENAME TO calcutta_invitations;

-- =============================================================================
-- STEP 12: Recreate primary keys
-- =============================================================================

ALTER TABLE core.calcuttas ADD CONSTRAINT calcuttas_pkey PRIMARY KEY (id);
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_pkey PRIMARY KEY (id);
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_pkey PRIMARY KEY (id);
ALTER TABLE core.entries ADD CONSTRAINT entries_pkey PRIMARY KEY (id);
ALTER TABLE core.entry_teams ADD CONSTRAINT entry_teams_pkey PRIMARY KEY (id);

-- =============================================================================
-- STEP 13: Recreate unique constraints
-- =============================================================================

ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT uq_core_calcutta_scoring_rules UNIQUE (calcutta_id, win_index);
ALTER TABLE core.payouts ADD CONSTRAINT uq_core_payouts_calcutta_position UNIQUE (calcutta_id, "position");

-- =============================================================================
-- STEP 14: Recreate check constraints
-- =============================================================================

ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_budget_positive CHECK (budget_points > 0);
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_max_bid_points_le_budget CHECK (max_bid_points <= budget_points);
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_max_bid_points_positive CHECK (max_bid_points > 0);
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_max_teams_gte_min CHECK (max_teams >= min_teams);
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_min_teams CHECK (min_teams >= 1);
ALTER TABLE core.calcuttas ADD CONSTRAINT ck_core_calcuttas_visibility CHECK (visibility = ANY (ARRAY['public'::text, 'unlisted'::text, 'private'::text]));
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT ck_calcutta_invitations_status CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'revoked'::text]));
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT chk_scoring_rules_points_nonneg CHECK (points_awarded >= 0);
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT chk_scoring_rules_win_index_nonneg CHECK (win_index >= 0);
ALTER TABLE core.entry_teams ADD CONSTRAINT chk_entry_teams_bid_positive CHECK (bid_points > 0);

-- =============================================================================
-- STEP 15: Recreate indexes
-- =============================================================================

CREATE INDEX idx_calcutta_invitations_active ON core.calcutta_invitations USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcutta_invitations_calcutta_id_active ON core.calcutta_invitations USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcutta_invitations_invited_by ON core.calcutta_invitations USING btree (invited_by);
CREATE INDEX idx_calcutta_invitations_user_id ON core.calcutta_invitations USING btree (user_id);
CREATE UNIQUE INDEX uq_calcutta_invitations_calcutta_user ON core.calcutta_invitations USING btree (calcutta_id, user_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcutta_scoring_rules_active ON core.calcutta_scoring_rules USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id ON core.calcutta_scoring_rules USING btree (calcutta_id);
CREATE INDEX idx_calcuttas_active ON core.calcuttas USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_calcuttas_created_by ON core.calcuttas USING btree (created_by);
CREATE INDEX idx_core_calcuttas_budget_points ON core.calcuttas USING btree (budget_points);
CREATE INDEX idx_core_calcuttas_owner_id ON core.calcuttas USING btree (owner_id);
CREATE INDEX idx_core_calcuttas_tournament_id ON core.calcuttas USING btree (tournament_id);
CREATE INDEX idx_entries_active ON core.entries USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_entries_calcutta_id_active ON core.entries USING btree (calcutta_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_entries_user_id ON core.entries USING btree (user_id);
CREATE UNIQUE INDEX uq_entries_user_calcutta ON core.entries USING btree (user_id, calcutta_id) WHERE ((user_id IS NOT NULL) AND (deleted_at IS NULL));
CREATE INDEX idx_entry_teams_active ON core.entry_teams USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_entry_teams_entry_id_active ON core.entry_teams USING btree (entry_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_entry_teams_team_id_active ON core.entry_teams USING btree (team_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uq_core_entry_teams_entry_team ON core.entry_teams USING btree (entry_id, team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_payouts_active ON core.payouts USING btree (id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_core_payouts_calcutta_id ON core.payouts USING btree (calcutta_id);

-- =============================================================================
-- STEP 16: Recreate foreign keys
-- =============================================================================

ALTER TABLE core.calcuttas ADD CONSTRAINT calcuttas_created_by_fkey FOREIGN KEY (created_by) REFERENCES core.users(id);
ALTER TABLE core.calcuttas ADD CONSTRAINT calcuttas_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES core.users(id);
ALTER TABLE core.calcuttas ADD CONSTRAINT calcuttas_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id);
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES core.users(id);
ALTER TABLE core.calcutta_invitations ADD CONSTRAINT calcutta_invitations_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT calcutta_scoring_rules_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE core.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE core.entries ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id);
ALTER TABLE core.entry_teams ADD CONSTRAINT entry_teams_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES core.entries(id) ON DELETE CASCADE;
ALTER TABLE core.entry_teams ADD CONSTRAINT entry_teams_team_id_fkey FOREIGN KEY (team_id) REFERENCES core.teams(id);
ALTER TABLE core.payouts ADD CONSTRAINT payouts_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;
ALTER TABLE lab.entries ADD CONSTRAINT entries_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_calcutta_id_fkey FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id);
ALTER TABLE lab.pipeline_calcutta_runs ADD CONSTRAINT pipeline_calcutta_runs_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES lab.entries(id);

-- =============================================================================
-- STEP 17: Recreate triggers
-- =============================================================================

CREATE TRIGGER trg_calcuttas_immutable_created_by BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.immutable_created_by();
CREATE TRIGGER trg_core_calcuttas_updated_at BEFORE UPDATE ON core.calcuttas FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_invitations_updated_at BEFORE UPDATE ON core.calcutta_invitations FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_calcutta_scoring_rules_updated_at BEFORE UPDATE ON core.calcutta_scoring_rules FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entries_updated_at BEFORE UPDATE ON core.entries FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
CREATE TRIGGER trg_core_entry_teams_updated_at BEFORE UPDATE ON core.entry_teams FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- STEP 18: Recreate derived views with original names
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
