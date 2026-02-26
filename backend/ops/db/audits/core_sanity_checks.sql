-- Sanity checks for Phase 3 backfill into core.*
-- This is intentionally lightweight and read-only.

-- Public vs core row counts
SELECT 'tournaments' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.tournaments WHERE deleted_at IS NULL) AS core_count;
SELECT 'schools' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.schools WHERE deleted_at IS NULL) AS core_count;
SELECT 'teams' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.teams WHERE deleted_at IS NULL) AS core_count;
SELECT 'kenpom_stats' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.team_kenpom_stats WHERE deleted_at IS NULL) AS core_count;
SELECT 'pools' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.pools WHERE deleted_at IS NULL) AS core_count;
SELECT 'portfolios' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.portfolios WHERE deleted_at IS NULL) AS core_count;
SELECT 'investments' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.investments WHERE deleted_at IS NULL) AS core_count;
SELECT 'payouts' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.payouts WHERE deleted_at IS NULL) AS core_count;
SELECT 'scoring_rules' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.pool_scoring_rules WHERE deleted_at IS NULL) AS core_count;

-- Seasons should cover all tournaments (we only backfilled tournaments with a computed year)
SELECT 'seasons' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.seasons) AS core_count;

-- Basic FK sanity in core
SELECT 'orphan_core_teams_school' AS metric, COUNT(*) AS n
FROM core.teams t
LEFT JOIN core.schools s ON s.id = t.school_id
WHERE t.deleted_at IS NULL AND s.id IS NULL;

SELECT 'orphan_core_teams_tournament' AS metric, COUNT(*) AS n
FROM core.teams t
LEFT JOIN core.tournaments tr ON tr.id = t.tournament_id
WHERE t.deleted_at IS NULL AND tr.id IS NULL;

SELECT 'orphan_core_portfolios_pool' AS metric, COUNT(*) AS n
FROM core.portfolios p
LEFT JOIN core.pools c ON c.id = p.pool_id
WHERE p.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_core_investments_portfolio' AS metric, COUNT(*) AS n
FROM core.investments i
LEFT JOIN core.portfolios p ON p.id = i.portfolio_id
WHERE i.deleted_at IS NULL AND p.id IS NULL;

SELECT 'orphan_core_investments_team' AS metric, COUNT(*) AS n
FROM core.investments i
LEFT JOIN core.teams t ON t.id = i.team_id
WHERE i.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_core_payouts_pool' AS metric, COUNT(*) AS n
FROM core.payouts p
LEFT JOIN core.pools c ON c.id = p.pool_id
WHERE p.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_core_scoring_rules_pool' AS metric, COUNT(*) AS n
FROM core.pool_scoring_rules r
LEFT JOIN core.pools c ON c.id = r.pool_id
WHERE r.deleted_at IS NULL AND c.id IS NULL;
