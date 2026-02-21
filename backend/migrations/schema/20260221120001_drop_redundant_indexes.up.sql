-- Drop redundant index: team_id is already the PK on core.team_kenpom_stats
DROP INDEX IF EXISTS core.idx_core_team_kenpom_stats_team_id;

-- Drop useless index: nobody queries calcuttas by budget_points alone
DROP INDEX IF EXISTS core.idx_core_calcuttas_budget_points;
