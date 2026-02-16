-- Restore low-cardinality index
CREATE INDEX IF NOT EXISTS idx_core_calcuttas_budget_points ON core.calcuttas (budget_points);

-- Restore original trigger naming
DROP TRIGGER IF EXISTS trg_core_calcutta_invitations_updated_at ON core.calcutta_invitations;
CREATE TRIGGER set_updated_at
  BEFORE UPDATE ON core.calcutta_invitations
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Restore unused views (stubs - original definitions may vary)
-- These were unused so we don't recreate them in rollback

-- Restore orphaned trigger functions (stubs)
CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_entry_evaluation_request()
RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_market_share_run()
RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION derived.enqueue_run_job_for_strategy_generation_run()
RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;
