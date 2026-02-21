-- Reverse 2.4: Restore legacy index names

-- Re-create redundant index
CREATE INDEX IF NOT EXISTS idx_silver_simulated_tournaments_tournament_id ON derived.simulated_teams USING btree (tournament_id);

-- derived.predicted_game_outcomes
ALTER INDEX IF EXISTS idx_derived_predicted_game_outcomes_tournament_id
    RENAME TO idx_silver_predicted_game_outcomes_tournament_id;

-- derived.simulated_teams
ALTER INDEX IF EXISTS idx_derived_simulated_teams_team_id
    RENAME TO idx_silver_simulated_tournaments_team_id;
ALTER INDEX IF EXISTS idx_derived_simulated_teams_sim_id
    RENAME TO idx_silver_simulated_tournaments_sim_id;
ALTER INDEX IF EXISTS idx_derived_simulated_teams_batch_id
    RENAME TO idx_analytics_simulated_tournaments_batch_id;
ALTER INDEX IF EXISTS uq_derived_simulated_teams_batch_sim_team
    RENAME TO uq_analytics_simulated_tournaments_batch_sim_team;

-- derived.simulated_tournaments
ALTER INDEX IF EXISTS idx_derived_simulated_tournaments_tournament_id
    RENAME TO idx_analytics_tournament_simulation_batches_tournament_id;
ALTER INDEX IF EXISTS idx_derived_simulated_tournaments_state_id
    RENAME TO idx_analytics_tournament_simulation_batches_snapshot_id;
ALTER INDEX IF EXISTS uq_derived_simulated_tournaments_natural_key
    RENAME TO uq_analytics_tournament_simulation_batches_natural_key;

-- derived.simulation_state_teams
ALTER INDEX IF EXISTS idx_derived_simulation_state_teams_team_id
    RENAME TO idx_analytics_tournament_state_snapshot_teams_team_id;
ALTER INDEX IF EXISTS idx_derived_simulation_state_teams_state_id
    RENAME TO idx_analytics_tournament_state_snapshot_teams_snapshot_id;

-- derived.simulation_states
ALTER INDEX IF EXISTS idx_derived_simulation_states_tournament_id
    RENAME TO idx_analytics_tournament_state_snapshots_tournament_id;

-- Reverse 2.3: Restore ON DELETE CASCADE

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_team2_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey
    FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_team1_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey
    FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_tournament_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;

-- Reverse 2.2: Drop triggers

DROP TRIGGER IF EXISTS trg_lab_evaluation_entry_results_updated_at ON lab.evaluation_entry_results;
DROP TRIGGER IF EXISTS trg_derived_predicted_team_values_updated_at ON derived.predicted_team_values;
DROP TRIGGER IF EXISTS trg_derived_prediction_batches_updated_at ON derived.prediction_batches;

-- Reverse 2.1: Drop columns

ALTER TABLE lab.pipeline_calcutta_runs DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE lab.pipeline_runs DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE lab.evaluation_entry_results DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE lab.evaluation_entry_results DROP COLUMN IF EXISTS updated_at;
ALTER TABLE derived.run_jobs DROP COLUMN IF EXISTS deleted_at;
