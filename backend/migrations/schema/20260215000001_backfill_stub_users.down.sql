-- Revert backfill: reassign entries from stub users back to NULL, then delete stub users

-- 1. Unlink entries from stub users
UPDATE core.entries e
SET user_id = NULL
FROM core.users u
WHERE e.user_id = u.id
  AND u.status = 'stub'
  AND u.deleted_at IS NULL;

-- 2. Delete stub users
DELETE FROM core.users WHERE status = 'stub';
