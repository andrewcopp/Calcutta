-- name: ListEntriesByCalcuttaID :many
WITH entry_points AS (
    SELECT
        cp.entry_id,
        COALESCE(SUM(cpt.actual_points), 0)::float8 AS total_points
    FROM calcutta_portfolios cp
    LEFT JOIN calcutta_portfolio_teams cpt ON cp.id = cpt.portfolio_id
    WHERE cp.deleted_at IS NULL AND cpt.deleted_at IS NULL
    GROUP BY cp.entry_id
)
SELECT
    ce.id,
    ce.name,
    ce.user_id,
    ce.calcutta_id,
    ce.created_at,
    ce.updated_at,
    ce.deleted_at,
    COALESCE(ep.total_points, 0)::float8 AS total_points
FROM calcutta_entries ce
LEFT JOIN entry_points ep ON ce.id = ep.entry_id
WHERE ce.calcutta_id = $1 AND ce.deleted_at IS NULL
ORDER BY ep.total_points DESC NULLS LAST, ce.created_at DESC;

-- name: GetEntryByID :one
SELECT
    id,
    name,
    user_id,
    calcutta_id,
    created_at,
    updated_at,
    deleted_at
FROM calcutta_entries
WHERE id = $1 AND deleted_at IS NULL;
