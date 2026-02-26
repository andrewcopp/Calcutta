-- name: ListScoringRules :many
SELECT
    id,
    pool_id,
    win_index AS round,
    points_awarded AS points,
    created_at,
    updated_at
FROM core.pool_scoring_rules
WHERE pool_id = $1 AND deleted_at IS NULL
ORDER BY win_index ASC;

-- name: CreateScoringRule :exec
INSERT INTO core.pool_scoring_rules (id, pool_id, win_index, points_awarded, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);
