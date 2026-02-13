-- Prepare for archival of old Lab tables that are replaced by the new lab schema
-- The new lab schema (lab.investment_models, lab.entries, lab.evaluations)
-- replaces the overly complex derived.algorithms, derived.candidates,
-- derived.suite_*, derived.synthetic_calcutta_*, etc.
--
-- NOTE: This migration only creates the archive schema and drops FK constraints.
-- The actual table moves will happen in a future migration once the Go code
-- that references these tables is removed.

CREATE SCHEMA IF NOT EXISTS archive;

COMMENT ON SCHEMA archive IS 'For archived Lab tables replaced by new lab schema (investment_models, entries, evaluations). Tables will be moved here once Go code migration is complete.';

--------------------------------------------------------------------------------
-- Drop FK constraints to allow future archival without blocking
-- The referenced IDs are stored in params_json anyway, so we don't lose data
--------------------------------------------------------------------------------

-- simulation_cohorts (renamed from synthetic_calcutta_cohorts) references algorithms
ALTER TABLE IF EXISTS derived.simulation_cohorts
    DROP CONSTRAINT IF EXISTS synthetic_calcutta_cohorts_game_outcomes_algorithm_id_fkey;
ALTER TABLE IF EXISTS derived.simulation_cohorts
    DROP CONSTRAINT IF EXISTS simulation_cohorts_game_outcomes_algorithm_id_fkey;
ALTER TABLE IF EXISTS derived.simulation_cohorts
    DROP CONSTRAINT IF EXISTS synthetic_calcutta_cohorts_market_share_algorithm_id_fkey;
ALTER TABLE IF EXISTS derived.simulation_cohorts
    DROP CONSTRAINT IF EXISTS simulation_cohorts_market_share_algorithm_id_fkey;

-- simulation_runs references market_share_runs and strategy_generation_runs
ALTER TABLE IF EXISTS derived.simulation_runs
    DROP CONSTRAINT IF EXISTS simulation_runs_market_share_run_id_fkey;
ALTER TABLE IF EXISTS derived.simulation_runs
    DROP CONSTRAINT IF EXISTS simulation_runs_strategy_generation_run_id_fkey;

-- synthetic_calcuttas (old table, if exists) references strategy_generation_runs
ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    DROP CONSTRAINT IF EXISTS synthetic_calcuttas_focus_strategy_generation_run_id_fkey;

-- game_outcome_runs references algorithms
ALTER TABLE IF EXISTS derived.game_outcome_runs
    DROP CONSTRAINT IF EXISTS game_outcome_runs_algorithm_id_fkey;

-- market_share_runs references algorithms
ALTER TABLE IF EXISTS derived.market_share_runs
    DROP CONSTRAINT IF EXISTS market_share_runs_algorithm_id_fkey;

-- suite_calcutta_evaluations references strategy_generation_runs and market_share_runs
ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    DROP CONSTRAINT IF EXISTS suite_calcutta_evaluations_strategy_generation_run_id_fkey;
ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    DROP CONSTRAINT IF EXISTS suite_calcutta_evaluations_market_share_run_id_fkey;

-- simulated_entries references candidates
ALTER TABLE IF EXISTS derived.simulated_entries
    DROP CONSTRAINT IF EXISTS simulated_entries_source_candidate_id_fkey;

-- strategy_generation_runs references market_share_runs
ALTER TABLE IF EXISTS derived.strategy_generation_runs
    DROP CONSTRAINT IF EXISTS strategy_generation_runs_market_share_run_id_fkey;

-- candidates references strategy_generation_runs and market_share_runs
ALTER TABLE IF EXISTS derived.candidates
    DROP CONSTRAINT IF EXISTS candidates_strategy_generation_run_id_fkey;
ALTER TABLE IF EXISTS derived.candidates
    DROP CONSTRAINT IF EXISTS candidates_market_share_run_id_fkey;

--------------------------------------------------------------------------------
-- Tables to be archived in future migration (once Go code is removed):
--   derived.algorithms -> archive.algorithms
--   derived.candidates -> archive.candidates
--   derived.candidate_bids -> archive.candidate_bids
--   derived.synthetic_calcutta_candidates -> archive.synthetic_calcutta_candidates
--   derived.market_share_runs -> archive.market_share_runs
--   derived.strategy_generation_runs -> archive.strategy_generation_runs
--   derived.strategy_generation_run_bids -> archive.strategy_generation_run_bids
--   derived.suites -> archive.suites
--   derived.suite_scenarios -> archive.suite_scenarios
--   derived.suite_executions -> archive.suite_executions
--   derived.suite_calcutta_evaluations -> archive.suite_calcutta_evaluations
--   derived.entry_evaluation_requests -> archive.entry_evaluation_requests
--------------------------------------------------------------------------------
