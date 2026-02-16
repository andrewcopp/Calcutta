-- Drop orphaned trigger functions (created but never attached to triggers)
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_entry_evaluation_request();
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_market_share_run();
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_strategy_generation_run();

-- Drop unused views
DROP VIEW IF EXISTS derived.v_algorithms;
DROP VIEW IF EXISTS derived.v_strategy_generation_run_bids;

-- Fix trigger naming convention on calcutta_invitations
DROP TRIGGER IF EXISTS set_updated_at ON core.calcutta_invitations;
CREATE TRIGGER trg_core_calcutta_invitations_updated_at
  BEFORE UPDATE ON core.calcutta_invitations
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Drop low-cardinality index (budget_points has ~1 distinct value)
DROP INDEX IF EXISTS idx_core_calcuttas_budget_points;
