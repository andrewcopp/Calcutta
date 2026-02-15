-- Users: password hash
ALTER TABLE users
    ADD COLUMN password_hash TEXT;

-- Auth sessions: stored refresh token hash (rotatable)
CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    refresh_token_hash TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    user_agent TEXT,
    ip_address TEXT
);

CREATE UNIQUE INDEX uq_auth_sessions_refresh_token_hash ON auth_sessions(refresh_token_hash);
CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);
CREATE INDEX idx_auth_sessions_revoked_at ON auth_sessions(revoked_at);

-- Permissions and labels
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE label_permissions (
    label_id UUID NOT NULL REFERENCES labels(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (label_id, permission_id)
);

-- Grants: assign either a label or a permission to a user at a scope
CREATE TABLE grants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    scope_type TEXT NOT NULL,
    scope_id UUID,
    label_id UUID REFERENCES labels(id),
    permission_id UUID REFERENCES permissions(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT grants_scope_type_check CHECK (scope_type IN ('global', 'calcutta')),
    CONSTRAINT grants_scope_id_check CHECK (
        (scope_type = 'global' AND scope_id IS NULL) OR
        (scope_type <> 'global' AND scope_id IS NOT NULL)
    ),
    CONSTRAINT grants_subject_check CHECK (
        (label_id IS NOT NULL AND permission_id IS NULL) OR
        (label_id IS NULL AND permission_id IS NOT NULL)
    )
);

CREATE INDEX idx_grants_user_id ON grants(user_id);
CREATE INDEX idx_grants_scope ON grants(scope_type, scope_id);
CREATE INDEX idx_grants_revoked_at ON grants(revoked_at);

-- Seed default permissions and labels (idempotent)
INSERT INTO permissions (key, description)
VALUES
    ('tournament.game.write', 'Update tournament game results and winners'),
    ('calcutta.config.write', 'Create and configure a calcutta'),
    ('calcutta.invite.write', 'Invite and manage calcutta members'),
    ('calcutta.read', 'Read calcutta data')
ON CONFLICT (key) DO NOTHING;

INSERT INTO labels (key, description)
VALUES
    ('global_admin', 'All permissions (admin)'),
    ('tournament_operator', 'Tournament operations'),
    ('calcutta_owner', 'Manage a specific calcutta'),
    ('player', 'Participate in a calcutta')
ON CONFLICT (key) DO NOTHING;

-- Wire label permissions (idempotent)
INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('tournament.game.write', 'calcutta.config.write', 'calcutta.invite.write', 'calcutta.read')
WHERE l.key = 'global_admin'
ON CONFLICT DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('tournament.game.write', 'calcutta.read')
WHERE l.key = 'tournament_operator'
ON CONFLICT DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('calcutta.config.write', 'calcutta.invite.write', 'calcutta.read')
WHERE l.key = 'calcutta_owner'
ON CONFLICT DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('calcutta.read')
WHERE l.key = 'player'
ON CONFLICT DO NOTHING;
