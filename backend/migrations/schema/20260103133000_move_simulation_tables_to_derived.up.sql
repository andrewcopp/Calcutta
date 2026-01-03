CREATE SCHEMA IF NOT EXISTS derived;

ALTER TABLE IF EXISTS analytics.tournament_state_snapshots SET SCHEMA derived;
ALTER TABLE IF EXISTS derived.tournament_state_snapshots RENAME TO simulation_states;

ALTER TABLE IF EXISTS analytics.tournament_state_snapshot_teams SET SCHEMA derived;
ALTER TABLE IF EXISTS derived.tournament_state_snapshot_teams RENAME TO simulation_state_teams;

ALTER TABLE IF EXISTS derived.simulation_state_teams
    RENAME COLUMN tournament_state_snapshot_id TO simulation_state_id;

ALTER TABLE IF EXISTS analytics.tournament_simulation_batches SET SCHEMA derived;
ALTER TABLE IF EXISTS derived.tournament_simulation_batches RENAME TO simulated_tournaments;

ALTER TABLE IF EXISTS analytics.simulated_tournaments
    RENAME TO simulated_teams;
ALTER TABLE IF EXISTS analytics.simulated_teams SET SCHEMA derived;

ALTER TABLE IF EXISTS derived.simulated_tournaments
    RENAME COLUMN tournament_state_snapshot_id TO simulation_state_id;

ALTER TABLE IF EXISTS derived.simulated_teams
    RENAME COLUMN tournament_simulation_batch_id TO simulated_tournament_id;

ALTER TABLE IF EXISTS analytics.calcutta_evaluation_runs SET SCHEMA derived;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    RENAME COLUMN tournament_simulation_batch_id TO simulated_tournament_id;

ALTER TABLE IF EXISTS analytics.entry_simulation_outcomes SET SCHEMA derived;
ALTER TABLE IF EXISTS analytics.entry_performance SET SCHEMA derived;

ALTER TABLE IF EXISTS lab_gold.strategy_generation_runs
    RENAME COLUMN tournament_simulation_batch_id TO simulated_tournament_id;
