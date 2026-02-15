-- Calcutta-level invitations
CREATE TABLE core.calcutta_invitations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_id uuid NOT NULL REFERENCES core.calcuttas(id),
    user_id uuid NOT NULL REFERENCES core.users(id),
    invited_by uuid NOT NULL REFERENCES core.users(id),
    status text NOT NULL DEFAULT 'pending',
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT ck_calcutta_invitations_status CHECK (status IN ('pending', 'accepted', 'declined'))
);
CREATE UNIQUE INDEX uq_calcutta_invitations_calcutta_user
    ON core.calcutta_invitations(calcutta_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_calcutta_invitations_calcutta_id ON core.calcutta_invitations(calcutta_id);
CREATE INDEX idx_calcutta_invitations_user_id ON core.calcutta_invitations(user_id);
CREATE INDEX idx_calcutta_invitations_invited_by ON core.calcutta_invitations(invited_by);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON core.calcutta_invitations
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Pool state controls
ALTER TABLE core.calcuttas ADD COLUMN bidding_open boolean NOT NULL DEFAULT true;
ALTER TABLE core.calcuttas ADD COLUMN bidding_locked_at timestamptz;
