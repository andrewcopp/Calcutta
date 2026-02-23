-- Rollback: derive_is_eliminated
-- Created: 2026-02-23 16:33:17 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Reset all teams to is_eliminated = false (the previous state).
UPDATE core.teams
SET is_eliminated = false, updated_at = NOW()
WHERE deleted_at IS NULL;
