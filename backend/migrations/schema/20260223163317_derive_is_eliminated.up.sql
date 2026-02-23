-- Migration: derive_is_eliminated
-- Created: 2026-02-23 16:33:17 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- In single elimination, if any team in a tournament has reached progress M
-- (wins + byes), all teams with progress < M have been eliminated.
-- Previously all teams had is_eliminated = false regardless of actual status.
UPDATE core.teams t
SET is_eliminated = true, updated_at = NOW()
WHERE t.deleted_at IS NULL
  AND (t.wins + t.byes) < (
    SELECT MAX(t2.wins + t2.byes)
    FROM core.teams t2
    WHERE t2.tournament_id = t.tournament_id AND t2.deleted_at IS NULL
  )
  AND (
    SELECT MAX(t2.wins + t2.byes)
    FROM core.teams t2
    WHERE t2.tournament_id = t.tournament_id AND t2.deleted_at IS NULL
  ) > 0;
