-- name: CreatePoolInvitation :exec
INSERT INTO core.pool_invitations (id, pool_id, user_id, invited_by, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW());

-- name: RevokePoolInvitation :execrows
UPDATE core.pool_invitations
SET status = 'revoked',
    revoked_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND status = 'pending';

-- name: ListPoolInvitationsByPoolID :many
SELECT id, pool_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.pool_invitations
WHERE pool_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetPoolInvitationByPoolAndUser :one
SELECT id, pool_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.pool_invitations
WHERE pool_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetPendingPoolInvitationByPoolAndUser :one
SELECT id, pool_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.pool_invitations
WHERE pool_id = $1 AND user_id = $2 AND status = 'pending' AND deleted_at IS NULL;

-- name: ListPendingInvitationsByUserID :many
SELECT id, pool_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.pool_invitations
WHERE user_id = $1 AND status = 'pending' AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: AcceptPoolInvitation :execrows
UPDATE core.pool_invitations
SET status = 'accepted',
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
