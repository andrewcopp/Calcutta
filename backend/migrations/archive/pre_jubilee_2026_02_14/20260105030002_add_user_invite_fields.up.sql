-- Users: invite / claim fields
ALTER TABLE users
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN invite_token_hash TEXT,
    ADD COLUMN invite_expires_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN invite_consumed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN invited_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN last_invite_sent_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE users
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'invited', 'requires_password_setup'));

CREATE UNIQUE INDEX uq_users_invite_token_hash ON users(invite_token_hash) WHERE invite_token_hash IS NOT NULL;
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_invite_expires_at ON users(invite_expires_at);
