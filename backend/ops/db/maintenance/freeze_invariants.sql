\pset pager off
\timing off

-- Baseline row counts (active rows)
SELECT 'tournaments' AS table, COUNT(*) AS n FROM core.tournaments WHERE deleted_at IS NULL;
SELECT 'tournament_teams' AS table, COUNT(*) AS n FROM core.teams WHERE deleted_at IS NULL;
SELECT 'tournament_games (derived)' AS table, 0::bigint AS n;
SELECT 'tournament_team_kenpom_stats' AS table, COUNT(*) AS n FROM core.team_kenpom_stats WHERE deleted_at IS NULL;

SELECT 'calcuttas' AS table, COUNT(*) AS n FROM core.calcuttas WHERE deleted_at IS NULL;
SELECT 'calcutta_entries' AS table, COUNT(*) AS n FROM core.entries WHERE deleted_at IS NULL;
SELECT 'calcutta_entry_teams' AS table, COUNT(*) AS n FROM core.entry_teams WHERE deleted_at IS NULL;
SELECT 'calcutta_payouts' AS table, COUNT(*) AS n FROM core.payouts WHERE deleted_at IS NULL;
SELECT 'calcutta_rounds' AS table, COUNT(*) AS n FROM core.calcutta_scoring_rules WHERE deleted_at IS NULL;

-- Orphans (should be 0)
SELECT 'orphan_tournament_teams' AS check, COUNT(*) AS n
FROM core.teams tt
LEFT JOIN core.tournaments t ON t.id = tt.tournament_id
WHERE tt.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_kenpom_stats' AS check, COUNT(*) AS n
FROM core.team_kenpom_stats kps
LEFT JOIN core.teams tt ON tt.id = kps.team_id
WHERE kps.deleted_at IS NULL AND tt.id IS NULL;

SELECT 'orphan_calcuttas_tournament' AS check, COUNT(*) AS n
FROM core.calcuttas c
LEFT JOIN core.tournaments t ON t.id = c.tournament_id
WHERE c.deleted_at IS NULL AND t.id IS NULL;

SELECT 'orphan_calcuttas_owner' AS check, COUNT(*) AS n
FROM core.calcuttas c
LEFT JOIN core.users u ON u.id = c.owner_id
WHERE c.deleted_at IS NULL AND u.id IS NULL;

SELECT 'orphan_entries' AS check, COUNT(*) AS n
FROM core.entries ce
LEFT JOIN core.calcuttas c ON c.id = ce.calcutta_id
WHERE ce.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_entry_teams_entry' AS check, COUNT(*) AS n
FROM core.entry_teams cet
LEFT JOIN core.entries ce ON ce.id = cet.entry_id
WHERE cet.deleted_at IS NULL AND ce.id IS NULL;

SELECT 'orphan_entry_teams_team' AS check, COUNT(*) AS n
FROM core.entry_teams cet
LEFT JOIN core.teams tt ON tt.id = cet.team_id
WHERE cet.deleted_at IS NULL AND tt.id IS NULL;

SELECT 'orphan_payouts' AS check, COUNT(*) AS n
FROM core.payouts cp
LEFT JOIN core.calcuttas c ON c.id = cp.calcutta_id
WHERE cp.deleted_at IS NULL AND c.id IS NULL;

SELECT 'orphan_rounds' AS check, COUNT(*) AS n
FROM core.calcutta_scoring_rules cr
LEFT JOIN core.calcuttas c ON c.id = cr.calcutta_id
WHERE cr.deleted_at IS NULL AND c.id IS NULL;

-- Useful distributions
SELECT
  'entries_per_calcutta' AS metric,
  ce.calcutta_id,
  COUNT(*) AS n
FROM core.entries ce
WHERE ce.deleted_at IS NULL
GROUP BY ce.calcutta_id
ORDER BY n DESC;

SELECT
  'teams_per_tournament' AS metric,
  tt.tournament_id,
  COUNT(*) AS n
FROM core.teams tt
WHERE tt.deleted_at IS NULL
GROUP BY tt.tournament_id
ORDER BY n DESC;
