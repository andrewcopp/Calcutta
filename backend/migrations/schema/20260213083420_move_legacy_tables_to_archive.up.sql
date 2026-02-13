-- Move legacy Lab tables to archive schema
-- These tables are no longer used by Go code after the Lab consolidation migration
-- The new lab.* schema (investment_models, entries, evaluations) replaces them

--------------------------------------------------------------------------------
-- Move tables to archive schema
-- Using ALTER TABLE SET SCHEMA which is fast (metadata-only operation)
--------------------------------------------------------------------------------

-- algorithms (replaced by lab.investment_models)
ALTER TABLE IF EXISTS derived.algorithms SET SCHEMA archive;

-- candidates and candidate_bids (replaced by lab.entries)
ALTER TABLE IF EXISTS derived.candidate_bids SET SCHEMA archive;
ALTER TABLE IF EXISTS derived.candidates SET SCHEMA archive;

-- strategy generation runs and bids (integrated into lab.entries workflow)
ALTER TABLE IF EXISTS derived.strategy_generation_run_bids SET SCHEMA archive;
ALTER TABLE IF EXISTS derived.strategy_generation_runs SET SCHEMA archive;

-- market share runs (integrated into lab workflow)
ALTER TABLE IF EXISTS derived.market_share_runs SET SCHEMA archive;

-- game outcome runs (kept for now, may archive later)
-- ALTER TABLE IF EXISTS derived.game_outcome_runs SET SCHEMA archive;

-- suite tables (replaced by lab.evaluations)
ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations SET SCHEMA archive;
ALTER TABLE IF EXISTS derived.suite_executions SET SCHEMA archive;
ALTER TABLE IF EXISTS derived.suite_scenarios SET SCHEMA archive;
ALTER TABLE IF EXISTS derived.suites SET SCHEMA archive;

-- synthetic calcutta candidates (replaced by lab.entries)
ALTER TABLE IF EXISTS derived.synthetic_calcutta_candidates SET SCHEMA archive;

-- synthetic calcuttas (old table, replaced by simulated_calcuttas)
ALTER TABLE IF EXISTS derived.synthetic_calcuttas SET SCHEMA archive;

-- entry evaluation requests (legacy workflow)
ALTER TABLE IF EXISTS derived.entry_evaluation_requests SET SCHEMA archive;

--------------------------------------------------------------------------------
-- Tables that REMAIN in derived.* (shared simulation infrastructure):
--   - simulation_runs
--   - simulation_cohorts
--   - simulated_calcuttas
--   - simulated_entries
--   - run_jobs
--   - run_artifacts
--   - game_outcome_runs (still used by Python runners)
--------------------------------------------------------------------------------
