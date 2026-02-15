-- Rollback: Move archived tables back to derived schema

ALTER TABLE IF EXISTS archive.algorithms SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.candidate_bids SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.candidates SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.strategy_generation_run_bids SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.strategy_generation_runs SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.market_share_runs SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.suite_calcutta_evaluations SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.suite_executions SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.suite_scenarios SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.suites SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.synthetic_calcutta_candidates SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.synthetic_calcuttas SET SCHEMA derived;
ALTER TABLE IF EXISTS archive.entry_evaluation_requests SET SCHEMA derived;
