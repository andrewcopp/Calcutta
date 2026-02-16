-- Recreate the dropped indexes.
CREATE INDEX IF NOT EXISTS idx_calcutta_invitations_calcutta_id
    ON core.calcutta_invitations(calcutta_id);

CREATE INDEX IF NOT EXISTS idx_calcutta_invitations_calcutta_user_active
    ON core.calcutta_invitations (calcutta_id, user_id) WHERE deleted_at IS NULL;
