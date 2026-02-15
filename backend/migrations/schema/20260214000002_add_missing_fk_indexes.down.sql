-- Remove FK indexes added in up migration.

-- Lab
DROP INDEX IF EXISTS lab.idx_lab_pipeline_calcutta_runs_evaluation_id;
DROP INDEX IF EXISTS lab.idx_lab_pipeline_calcutta_runs_entry_id;
DROP INDEX IF EXISTS lab.idx_lab_evaluations_simulated_calcutta_id;

-- Derived
DROP INDEX IF EXISTS derived.idx_derived_simulation_runs_simulation_run_batch_id;
DROP INDEX IF EXISTS derived.idx_derived_simulation_runs_calcutta_evaluation_run_id;
DROP INDEX IF EXISTS derived.idx_derived_simulation_runs_game_outcome_run_id;
DROP INDEX IF EXISTS derived.idx_derived_simulated_entries_source_entry_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_market_share_tournament_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_market_share_team_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_market_share_calcutta_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_game_outcomes_run_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_game_outcomes_team2_id;
DROP INDEX IF EXISTS derived.idx_derived_predicted_game_outcomes_team1_id;
DROP INDEX IF EXISTS derived.idx_derived_optimization_runs_calcutta_id;
DROP INDEX IF EXISTS derived.idx_derived_detailed_investment_report_team_id;
DROP INDEX IF EXISTS derived.idx_derived_game_outcome_runs_prediction_model_id;

-- Core
DROP INDEX IF EXISTS core.idx_calcutta_snapshot_entry_teams_team_id;
DROP INDEX IF EXISTS core.idx_calcutta_snapshot_entries_entry_id;
DROP INDEX IF EXISTS core.idx_label_permissions_permission_id;
DROP INDEX IF EXISTS core.idx_label_permissions_label_id;
DROP INDEX IF EXISTS core.idx_grants_permission_id;
DROP INDEX IF EXISTS core.idx_grants_label_id;
