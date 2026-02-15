-- Add indexes on foreign key columns that were missing indexes.
-- These prevent sequential scans on joins and cascading deletes.

-- Core
CREATE INDEX IF NOT EXISTS idx_grants_label_id ON core.grants(label_id);
CREATE INDEX IF NOT EXISTS idx_grants_permission_id ON core.grants(permission_id);
CREATE INDEX IF NOT EXISTS idx_label_permissions_label_id ON core.label_permissions(label_id);
CREATE INDEX IF NOT EXISTS idx_label_permissions_permission_id ON core.label_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_calcutta_snapshot_entries_entry_id ON core.calcutta_snapshot_entries(entry_id);
CREATE INDEX IF NOT EXISTS idx_calcutta_snapshot_entry_teams_team_id ON core.calcutta_snapshot_entry_teams(team_id);

-- Derived
CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_prediction_model_id ON derived.game_outcome_runs(prediction_model_id);
CREATE INDEX IF NOT EXISTS idx_derived_detailed_investment_report_team_id ON derived.detailed_investment_report(team_id);
CREATE INDEX IF NOT EXISTS idx_derived_optimization_runs_calcutta_id ON derived.optimization_runs(calcutta_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_game_outcomes_team1_id ON derived.predicted_game_outcomes(team1_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_game_outcomes_team2_id ON derived.predicted_game_outcomes(team2_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_game_outcomes_run_id ON derived.predicted_game_outcomes(run_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_market_share_calcutta_id ON derived.predicted_market_share(calcutta_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_market_share_team_id ON derived.predicted_market_share(team_id);
CREATE INDEX IF NOT EXISTS idx_derived_predicted_market_share_tournament_id ON derived.predicted_market_share(tournament_id);
CREATE INDEX IF NOT EXISTS idx_derived_simulated_entries_source_entry_id ON derived.simulated_entries(source_entry_id);
CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_game_outcome_run_id ON derived.simulation_runs(game_outcome_run_id);
CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_calcutta_evaluation_run_id ON derived.simulation_runs(calcutta_evaluation_run_id);
CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_simulation_run_batch_id ON derived.simulation_runs(simulation_run_batch_id);

-- Lab
CREATE INDEX IF NOT EXISTS idx_lab_evaluations_simulated_calcutta_id ON lab.evaluations(simulated_calcutta_id);
CREATE INDEX IF NOT EXISTS idx_lab_pipeline_calcutta_runs_entry_id ON lab.pipeline_calcutta_runs(entry_id);
CREATE INDEX IF NOT EXISTS idx_lab_pipeline_calcutta_runs_evaluation_id ON lab.pipeline_calcutta_runs(evaluation_id);
