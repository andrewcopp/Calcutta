-- name: ListCalcuttaRounds :many
SELECT id, calcutta_id, round, points, created_at, updated_at
FROM calcutta_rounds
WHERE calcutta_id = $1 AND deleted_at IS NULL
ORDER BY round ASC;

-- name: CreateCalcuttaRound :exec
INSERT INTO calcutta_rounds (id, calcutta_id, round, points, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);
