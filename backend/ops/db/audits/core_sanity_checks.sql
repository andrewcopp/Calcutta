-- Sanity checks for Phase 3 backfill into core.*
-- This is intentionally lightweight and read-only.

-- Public vs core row counts
SELECT 'tournaments' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.tournaments WHERE deleted_at IS NULL) AS core_count;
SELECT 'schools' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.schools WHERE deleted_at IS NULL) AS core_count;
SELECT 'teams' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.teams WHERE deleted_at IS NULL) AS core_count;
SELECT 'kenpom_stats' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.team_kenpom_stats WHERE deleted_at IS NULL) AS core_count;
SELECT 'calcuttas' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.calcuttas WHERE deleted_at IS NULL) AS core_count;
SELECT 'entries' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.entries WHERE deleted_at IS NULL) AS core_count;
SELECT 'entry_teams' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.entry_teams WHERE deleted_at IS NULL) AS core_count;
SELECT 'payouts' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.payouts WHERE deleted_at IS NULL) AS core_count;
SELECT 'scoring_rules' AS metric, NULL::bigint AS public_count, (SELECT COUNT(*) FROM core.calcutta_scoring_rules WHERE deleted_at IS NULL) AS core_count;

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

SELECT 'orphan_core_entries_calcutta' AS metric, COUNT(*) AS n
FROM core.entries e
LEFT JOIN core.calcuttas c ON c.id = e.calcutta_id
WHERE e.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_core_entry_teams_entry' AS metric, COUNT(*) AS n
FROM core.entry_teams et
LEFT JOIN core.entries e ON e.id = et.entry_id
WHERE et.deleted_at IS NULL AND e.id IS NULL;

SELECT 'orphan_core_entry_teams_team' AS metric, COUNT(*) AS n
FROM core.entry_teams et
LEFT JOIN core.teams t ON t.id = et.team_id
WHERE et.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_core_payouts_calcutta' AS metric, COUNT(*) AS n
FROM core.payouts p
LEFT JOIN core.calcuttas c ON c.id = p.calcutta_id
WHERE p.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_core_scoring_rules_calcutta' AS metric, COUNT(*) AS n
FROM core.calcutta_scoring_rules r
LEFT JOIN core.calcuttas c ON c.id = r.calcutta_id
WHERE r.deleted_at IS NULL AND c.id IS NULL;
