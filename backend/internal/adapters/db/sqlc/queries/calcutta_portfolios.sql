-- name: GetPortfolioByID :one
SELECT
    id,
    entry_id,
    maximum_points::float8 AS maximum_points,
    created_at::timestamptz AS created_at,
    updated_at::timestamptz AS updated_at,
    deleted_at::timestamptz AS deleted_at
FROM derived.portfolios
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPortfoliosByEntryID :many
SELECT
    id,
    entry_id,
    created_at::timestamptz AS created_at,
    updated_at::timestamptz AS updated_at,
    deleted_at::timestamptz AS deleted_at
FROM derived.portfolios
WHERE entry_id = $1 AND deleted_at IS NULL;
