-- name: GetPortfolioByID :one
SELECT
    id,
    entry_id,
    maximum_points::float8 AS maximum_points,
    created_at,
    updated_at,
    deleted_at
FROM calcutta_portfolios
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPortfoliosByEntryID :many
SELECT id, entry_id, created_at, updated_at, deleted_at
FROM calcutta_portfolios
WHERE entry_id = $1 AND deleted_at IS NULL;

-- name: ListPortfolios :many
SELECT
    id,
    entry_id,
    maximum_points::float8 AS maximum_points,
    created_at,
    updated_at,
    deleted_at
FROM calcutta_portfolios
WHERE entry_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreatePortfolio :exec
INSERT INTO calcutta_portfolios (id, entry_id, maximum_points, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePortfolio :execrows
UPDATE calcutta_portfolios
SET maximum_points = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;
