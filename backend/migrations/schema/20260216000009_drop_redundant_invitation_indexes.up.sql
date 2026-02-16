-- Drop redundant indexes on core.calcutta_invitations.
--
-- idx_calcutta_invitations_calcutta_id is superseded by the partial
-- idx_calcutta_invitations_calcutta_id_active (WHERE deleted_at IS NULL).
--
-- idx_calcutta_invitations_calcutta_user_active is redundant with the
-- unique index uq_calcutta_invitations_calcutta_user which covers
-- the same (calcutta_id, user_id) WHERE deleted_at IS NULL.

DROP INDEX IF EXISTS core.idx_calcutta_invitations_calcutta_id;
DROP INDEX IF EXISTS core.idx_calcutta_invitations_calcutta_user_active;
