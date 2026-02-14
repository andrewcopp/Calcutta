ALTER TABLE IF EXISTS lab_gold.strategy_generation_runs
    RENAME COLUMN simulated_tournament_id TO tournament_simulation_batch_id;

ALTER TABLE IF EXISTS derived.entry_simulation_outcomes SET SCHEMA analytics;
ALTER TABLE IF EXISTS derived.entry_performance SET SCHEMA analytics;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs
    RENAME COLUMN simulated_tournament_id TO tournament_simulation_batch_id;

ALTER TABLE IF EXISTS derived.calcutta_evaluation_runs SET SCHEMA analytics;

ALTER TABLE IF EXISTS derived.simulated_teams
    RENAME COLUMN simulated_tournament_id TO tournament_simulation_batch_id;

ALTER TABLE IF EXISTS derived.simulated_teams SET SCHEMA analytics;
ALTER TABLE IF EXISTS analytics.simulated_teams
    RENAME TO simulated_tournaments;

ALTER TABLE IF EXISTS derived.simulated_tournaments
    RENAME COLUMN simulation_state_id TO tournament_state_snapshot_id;

ALTER TABLE IF EXISTS derived.simulated_tournaments RENAME TO tournament_simulation_batches;
ALTER TABLE IF EXISTS derived.tournament_simulation_batches SET SCHEMA analytics;

ALTER TABLE IF EXISTS derived.simulation_state_teams
    RENAME COLUMN simulation_state_id TO tournament_state_snapshot_id;

ALTER TABLE IF EXISTS derived.simulation_state_teams RENAME TO tournament_state_snapshot_teams;
ALTER TABLE IF EXISTS derived.simulation_states RENAME TO tournament_state_snapshots;

ALTER TABLE IF EXISTS derived.tournament_state_snapshot_teams SET SCHEMA analytics;
ALTER TABLE IF EXISTS derived.tournament_state_snapshots SET SCHEMA analytics;

-- Leave schema derived in place; it may be used by other migrations.
