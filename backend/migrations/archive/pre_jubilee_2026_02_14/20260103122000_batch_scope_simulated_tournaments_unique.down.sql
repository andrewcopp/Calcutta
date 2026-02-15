-- Rollback: remove batch-scoped uniqueness and restore legacy uniqueness.

DROP INDEX IF EXISTS analytics.uq_analytics_simulated_tournaments_batch_sim_team;
DROP INDEX IF EXISTS analytics.uq_analytics_simulated_tournaments_legacy_sim_team;

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_simulated_tournaments_tournament_sim_team
ON analytics.simulated_tournaments (tournament_id, sim_id, team_id)
WHERE deleted_at IS NULL;
