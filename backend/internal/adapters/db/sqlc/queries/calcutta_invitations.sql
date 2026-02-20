-- name: CreateCalcuttaInvitation :exec
INSERT INTO core.calcutta_invitations (id, calcutta_id, user_id, invited_by, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW());

-- name: RevokeCalcuttaInvitation :execrows
UPDATE core.calcutta_invitations
SET status = 'revoked',
    revoked_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND status = 'pending';

-- name: ListCalcuttaInvitationsByCalcuttaID :many
SELECT id, calcutta_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.calcutta_invitations
WHERE calcutta_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCalcuttaInvitationByCalcuttaAndUser :one
SELECT id, calcutta_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.calcutta_invitations
WHERE calcutta_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetPendingCalcuttaInvitationByCalcuttaAndUser :one
SELECT id, calcutta_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.calcutta_invitations
WHERE calcutta_id = $1 AND user_id = $2 AND status = 'pending' AND deleted_at IS NULL;

-- name: ListPendingInvitationsByUserID :many
SELECT id, calcutta_id, user_id, invited_by, status, revoked_at, created_at, updated_at
FROM core.calcutta_invitations
WHERE user_id = $1 AND status = 'pending' AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: AcceptCalcuttaInvitation :execrows
UPDATE core.calcutta_invitations
SET status = 'accepted',
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
