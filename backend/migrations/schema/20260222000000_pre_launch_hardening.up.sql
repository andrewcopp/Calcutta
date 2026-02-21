-- 1A: Pre-launch database hardening

-- 1. Add CHECK constraint on calcuttas.name to prevent empty strings
ALTER TABLE core.calcuttas
    ADD CONSTRAINT ck_core_calcuttas_name_not_empty CHECK (length(trim(name)) > 0);

-- 2. Add expires_at column + index on idempotency_keys for TTL cleanup
ALTER TABLE core.idempotency_keys
    ADD COLUMN expires_at timestamp with time zone DEFAULT (NOW() + INTERVAL '24 hours');

CREATE INDEX idx_core_idempotency_keys_expires_at
    ON core.idempotency_keys (expires_at);

-- 3. Drop redundant index on calcutta_scoring_rules
-- The UNIQUE constraint uq_core_calcutta_scoring_rules (calcutta_id, win_index)
-- already provides a B-tree index with calcutta_id as the leading column.
DROP INDEX IF EXISTS idx_core_calcutta_scoring_rules_calcutta_id;

-- 4. Add partial index for pending invitation lookups
CREATE INDEX idx_calcutta_invitations_user_pending
    ON core.calcutta_invitations (user_id)
    WHERE status = 'pending' AND deleted_at IS NULL;

-- 5. Add index on entries.name for Hall of Fame queries
CREATE INDEX idx_core_entries_name
    ON core.entries (name)
    WHERE deleted_at IS NULL;
