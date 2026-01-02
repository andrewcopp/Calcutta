-- name: ListCalcuttaRounds :many
SELECT
    id,
    calcutta_id,
    win_index AS round,
    points_awarded AS points,
    created_at,
    updated_at
FROM core.calcutta_scoring_rules
WHERE calcutta_id = $1 AND deleted_at IS NULL
ORDER BY win_index ASC;

-- name: CreateCalcuttaRound :exec
INSERT INTO core.calcutta_scoring_rules (id, calcutta_id, win_index, points_awarded, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);
