CREATE SCHEMA IF NOT EXISTS compute;

ALTER TABLE derived.prediction_batches SET SCHEMA compute;
ALTER TABLE derived.predicted_team_values SET SCHEMA compute;
ALTER TABLE derived.game_outcome_runs SET SCHEMA compute;
ALTER TABLE derived.predicted_game_outcomes SET SCHEMA compute;
ALTER TABLE derived.tournament_snapshots SET SCHEMA compute;
ALTER TABLE derived.tournament_snapshot_teams SET SCHEMA compute;
ALTER TABLE derived.simulated_tournaments SET SCHEMA compute;
ALTER TABLE derived.simulated_teams SET SCHEMA compute;
