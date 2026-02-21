-- 2.1: Add missing deleted_at columns

ALTER TABLE derived.run_jobs ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE lab.evaluation_entry_results ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone DEFAULT now() NOT NULL;
ALTER TABLE lab.evaluation_entry_results ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE lab.pipeline_runs ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE lab.pipeline_calcutta_runs ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

-- 2.2: Add missing set_updated_at triggers

CREATE TRIGGER trg_derived_prediction_batches_updated_at
    BEFORE UPDATE ON derived.prediction_batches
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

CREATE TRIGGER trg_derived_predicted_team_values_updated_at
    BEFORE UPDATE ON derived.predicted_team_values
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

CREATE TRIGGER trg_lab_evaluation_entry_results_updated_at
    BEFORE UPDATE ON lab.evaluation_entry_results
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 2.3: Fix dangerous ON DELETE CASCADE on derived.predicted_game_outcomes
-- Drop existing FK constraints and re-add as RESTRICT

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_tournament_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE RESTRICT;

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_team1_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey
    FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE RESTRICT;

ALTER TABLE derived.predicted_game_outcomes DROP CONSTRAINT IF EXISTS predicted_game_outcomes_team2_id_fkey;
ALTER TABLE derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey
    FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE RESTRICT;

-- 2.4: Rename legacy index prefixes (analytics_ and silver_)

-- derived.simulation_states (was analytics_tournament_state_snapshots)
ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshots_tournament_id
    RENAME TO idx_derived_simulation_states_tournament_id;

-- derived.simulation_state_teams (was analytics_tournament_state_snapshot_teams)
ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshot_teams_snapshot_id
    RENAME TO idx_derived_simulation_state_teams_state_id;
ALTER INDEX IF EXISTS idx_analytics_tournament_state_snapshot_teams_team_id
    RENAME TO idx_derived_simulation_state_teams_team_id;

-- derived.simulated_tournaments (was analytics_tournament_simulation_batches)
ALTER INDEX IF EXISTS uq_analytics_tournament_simulation_batches_natural_key
    RENAME TO uq_derived_simulated_tournaments_natural_key;
ALTER INDEX IF EXISTS idx_analytics_tournament_simulation_batches_snapshot_id
    RENAME TO idx_derived_simulated_tournaments_state_id;
ALTER INDEX IF EXISTS idx_analytics_tournament_simulation_batches_tournament_id
    RENAME TO idx_derived_simulated_tournaments_tournament_id;

-- derived.simulated_teams (was analytics_simulated_tournaments / silver_simulated_tournaments)
ALTER INDEX IF EXISTS uq_analytics_simulated_tournaments_batch_sim_team
    RENAME TO uq_derived_simulated_teams_batch_sim_team;
ALTER INDEX IF EXISTS idx_analytics_simulated_tournaments_batch_id
    RENAME TO idx_derived_simulated_teams_batch_id;
ALTER INDEX IF EXISTS idx_silver_simulated_tournaments_sim_id
    RENAME TO idx_derived_simulated_teams_sim_id;
ALTER INDEX IF EXISTS idx_silver_simulated_tournaments_team_id
    RENAME TO idx_derived_simulated_teams_team_id;

-- derived.predicted_game_outcomes (was silver_predicted_game_outcomes)
ALTER INDEX IF EXISTS idx_silver_predicted_game_outcomes_tournament_id
    RENAME TO idx_derived_predicted_game_outcomes_tournament_id;

-- Drop redundant standalone tournament_id index on derived.simulated_teams
-- (already covered by composite index idx_derived_simulated_teams_sim_id which includes tournament_id)
DROP INDEX IF EXISTS idx_silver_simulated_tournaments_tournament_id;
