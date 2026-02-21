-- Rollback: pre-launch database hardening

DROP INDEX IF EXISTS idx_core_entries_name;

DROP INDEX IF EXISTS idx_calcutta_invitations_user_pending;

CREATE INDEX idx_core_calcutta_scoring_rules_calcutta_id
    ON core.calcutta_scoring_rules USING btree (calcutta_id);

DROP INDEX IF EXISTS idx_core_idempotency_keys_expires_at;

ALTER TABLE core.idempotency_keys
    DROP COLUMN IF EXISTS expires_at;

ALTER TABLE core.calcuttas
    DROP CONSTRAINT IF EXISTS ck_core_calcuttas_name_not_empty;
