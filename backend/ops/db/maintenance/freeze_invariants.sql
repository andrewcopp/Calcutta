\pset pager off
\timing off

-- Baseline row counts (active rows)
SELECT 'tournaments' AS table, COUNT(*) AS n FROM core.tournaments WHERE deleted_at IS NULL;
SELECT 'tournament_teams' AS table, COUNT(*) AS n FROM core.teams WHERE deleted_at IS NULL;
SELECT 'tournament_games (derived)' AS table, 0::bigint AS n;
SELECT 'tournament_team_kenpom_stats' AS table, COUNT(*) AS n FROM core.team_kenpom_stats WHERE deleted_at IS NULL;

SELECT 'pools' AS table, COUNT(*) AS n FROM core.pools WHERE deleted_at IS NULL;
SELECT 'portfolios' AS table, COUNT(*) AS n FROM core.portfolios WHERE deleted_at IS NULL;
SELECT 'investments' AS table, COUNT(*) AS n FROM core.investments WHERE deleted_at IS NULL;
SELECT 'payouts' AS table, COUNT(*) AS n FROM core.payouts WHERE deleted_at IS NULL;
SELECT 'scoring_rules' AS table, COUNT(*) AS n FROM core.pool_scoring_rules WHERE deleted_at IS NULL;

-- Orphans (should be 0)
SELECT 'orphan_tournament_teams' AS check, COUNT(*) AS n
FROM core.teams tt
LEFT JOIN core.tournaments t ON t.id = tt.tournament_id
WHERE tt.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_kenpom_stats' AS check, COUNT(*) AS n
FROM core.team_kenpom_stats kps
LEFT JOIN core.teams tt ON tt.id = kps.team_id
WHERE kps.deleted_at IS NULL AND tt.id IS NULL;

SELECT 'orphan_pools_tournament' AS check, COUNT(*) AS n
FROM core.pools p
LEFT JOIN core.tournaments t ON t.id = p.tournament_id
WHERE p.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_pools_owner' AS check, COUNT(*) AS n
FROM core.pools p
LEFT JOIN core.users u ON u.id = p.owner_id
WHERE p.deleted_at IS NULL AND u.id IS NULL;

SELECT 'orphan_portfolios' AS check, COUNT(*) AS n
FROM core.portfolios pf
LEFT JOIN core.pools p ON p.id = pf.pool_id
WHERE pf.deleted_at IS NULL AND p.id IS NULL;

SELECT 'orphan_investments_portfolio' AS check, COUNT(*) AS n
FROM core.investments i
LEFT JOIN core.portfolios pf ON pf.id = i.portfolio_id
WHERE i.deleted_at IS NULL AND pf.id IS NULL;

SELECT 'orphan_investments_team' AS check, COUNT(*) AS n
FROM core.investments i
LEFT JOIN core.teams tt ON tt.id = i.team_id
WHERE i.deleted_at IS NULL AND tt.id IS NULL;

SELECT 'orphan_payouts' AS check, COUNT(*) AS n
FROM core.payouts cp
LEFT JOIN core.pools p ON p.id = cp.pool_id
WHERE cp.deleted_at IS NULL AND p.id IS NULL;

SELECT 'orphan_scoring_rules' AS check, COUNT(*) AS n
FROM core.pool_scoring_rules sr
LEFT JOIN core.pools p ON p.id = sr.pool_id
WHERE sr.deleted_at IS NULL AND p.id IS NULL;

-- Useful distributions
SELECT
  'portfolios_per_pool' AS metric,
  pf.pool_id,
  COUNT(*) AS n
FROM core.portfolios pf
WHERE pf.deleted_at IS NULL
GROUP BY pf.pool_id
ORDER BY n DESC;

SELECT
  'teams_per_tournament' AS metric,
  tt.tournament_id,
  COUNT(*) AS n
FROM core.teams tt
WHERE tt.deleted_at IS NULL
GROUP BY tt.tournament_id
ORDER BY n DESC;
