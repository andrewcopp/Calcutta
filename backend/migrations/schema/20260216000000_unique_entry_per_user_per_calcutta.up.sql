-- Ensure each user can only have one active entry per Calcutta.
-- Partial index: historical null-user entries are exempt, soft-deleted rows are exempt.
CREATE UNIQUE INDEX uq_entries_user_calcutta
ON core.entries (user_id, calcutta_id)
WHERE user_id IS NOT NULL AND deleted_at IS NULL;
