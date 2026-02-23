ALTER TABLE compute.prediction_batches SET SCHEMA derived;
ALTER TABLE compute.predicted_team_values SET SCHEMA derived;
ALTER TABLE compute.game_outcome_runs SET SCHEMA derived;
ALTER TABLE compute.predicted_game_outcomes SET SCHEMA derived;
ALTER TABLE compute.tournament_snapshots SET SCHEMA derived;
ALTER TABLE compute.tournament_snapshot_teams SET SCHEMA derived;
ALTER TABLE compute.simulated_tournaments SET SCHEMA derived;
ALTER TABLE compute.simulated_teams SET SCHEMA derived;

DROP SCHEMA IF EXISTS compute;
