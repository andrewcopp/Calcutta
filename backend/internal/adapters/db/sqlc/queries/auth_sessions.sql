-- name: CreateAuthSession :one
INSERT INTO core.auth_sessions (user_id, refresh_token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetAuthSessionByID :one
SELECT id, user_id, refresh_token_hash, expires_at, revoked_at
FROM core.auth_sessions
WHERE id = $1;

-- name: GetAuthSessionByRefreshTokenHash :one
SELECT id, user_id, refresh_token_hash, expires_at, revoked_at
FROM core.auth_sessions
WHERE refresh_token_hash = $1;

-- name: RotateAuthSessionRefreshToken :exec
UPDATE core.auth_sessions
SET refresh_token_hash = $2,
    expires_at = $3,
    updated_at = NOW(),
    last_used_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: RevokeAuthSession :exec
UPDATE core.auth_sessions
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;
