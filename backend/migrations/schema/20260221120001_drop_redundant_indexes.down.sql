-- Recreate the dropped indexes
CREATE INDEX IF NOT EXISTS idx_core_team_kenpom_stats_team_id ON core.team_kenpom_stats USING btree (team_id);
CREATE INDEX IF NOT EXISTS idx_core_calcuttas_budget_points ON core.calcuttas USING btree (budget_points);
