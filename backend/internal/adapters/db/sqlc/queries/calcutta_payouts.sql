-- name: ListCalcuttaPayouts :many
SELECT id, calcutta_id, position, amount_cents, created_at, updated_at
FROM core.payouts
WHERE calcutta_id = $1 AND deleted_at IS NULL
ORDER BY position ASC;

-- name: SoftDeletePayoutsByCalcuttaID :execrows
UPDATE core.payouts
SET deleted_at = @deleted_at, updated_at = @updated_at
WHERE calcutta_id = @calcutta_id AND deleted_at IS NULL;

-- name: CreatePayout :exec
INSERT INTO core.payouts (id, calcutta_id, position, amount_cents, created_at, updated_at)
VALUES (@id, @calcutta_id, @position, @amount_cents, @created_at, @updated_at);
