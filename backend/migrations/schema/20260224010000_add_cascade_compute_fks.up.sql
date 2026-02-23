ALTER TABLE compute.predicted_team_values
  DROP CONSTRAINT predicted_team_values_prediction_batch_id_fkey;
ALTER TABLE compute.predicted_team_values
  ADD CONSTRAINT predicted_team_values_prediction_batch_id_fkey
  FOREIGN KEY (prediction_batch_id) REFERENCES compute.prediction_batches(id) ON DELETE CASCADE;

ALTER TABLE compute.simulated_teams
  DROP CONSTRAINT simulated_tournaments_tournament_simulation_batch_id_fkey;
ALTER TABLE compute.simulated_teams
  ADD CONSTRAINT simulated_tournaments_tournament_simulation_batch_id_fkey
  FOREIGN KEY (simulated_tournament_id) REFERENCES compute.simulated_tournaments(id) ON DELETE CASCADE;
