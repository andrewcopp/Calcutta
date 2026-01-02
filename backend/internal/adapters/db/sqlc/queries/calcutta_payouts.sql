-- name: ListCalcuttaPayouts :many
SELECT id, calcutta_id, position, amount_cents, created_at, updated_at
FROM core.payouts
WHERE calcutta_id = $1 AND deleted_at IS NULL
ORDER BY position ASC;
