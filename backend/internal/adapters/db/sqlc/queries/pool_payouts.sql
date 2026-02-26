-- name: ListPoolPayouts :many
SELECT id, pool_id, position, amount_cents, created_at, updated_at
FROM core.payouts
WHERE pool_id = $1 AND deleted_at IS NULL
ORDER BY position ASC;

-- name: SoftDeletePayoutsByPoolID :execrows
UPDATE core.payouts
SET deleted_at = @deleted_at, updated_at = @updated_at
WHERE pool_id = @pool_id AND deleted_at IS NULL;

-- name: CreatePayout :exec
INSERT INTO core.payouts (id, pool_id, position, amount_cents, created_at, updated_at)
VALUES (@id, @pool_id, @position, @amount_cents, @created_at, @updated_at);
