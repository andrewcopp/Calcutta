-- name: GetOwnershipSummaryByID :one
SELECT
    id,
    portfolio_id,
    maximum_returns::float8 AS maximum_returns,
    created_at::timestamptz AS created_at,
    updated_at::timestamptz AS updated_at,
    deleted_at::timestamptz AS deleted_at
FROM derived.ownership_summaries
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOwnershipSummariesByPortfolioID :many
SELECT
    id,
    portfolio_id,
    created_at::timestamptz AS created_at,
    updated_at::timestamptz AS updated_at,
    deleted_at::timestamptz AS deleted_at
FROM derived.ownership_summaries
WHERE portfolio_id = $1 AND deleted_at IS NULL;

-- name: ListOwnershipSummariesByPortfolioIDs :many
SELECT
    id,
    portfolio_id,
    created_at::timestamptz AS created_at,
    updated_at::timestamptz AS updated_at,
    deleted_at::timestamptz AS deleted_at
FROM derived.ownership_summaries
WHERE portfolio_id = ANY(@portfolio_ids::uuid[]) AND deleted_at IS NULL;
