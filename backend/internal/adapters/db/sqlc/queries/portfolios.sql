-- name: ListPortfoliosByPoolID :many
WITH portfolio_investments AS (
    SELECT
        inv.portfolio_id,
        p.pool_id,
        inv.team_id,
        inv.credits::float8 AS credits,
        SUM(inv.credits::float8) OVER (
            PARTITION BY p.pool_id, inv.team_id
        ) AS team_total_credits
    FROM core.investments inv
    JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
    WHERE inv.deleted_at IS NULL
),
portfolio_returns AS (
    SELECT
        p.id AS portfolio_id,
        COALESCE(
            SUM(
                CASE
                    WHEN pi.team_total_credits > 0 THEN
                        core.pool_returns_for_progress(p.pool_id, tt.wins, tt.byes)::float8
                        * (pi.credits / pi.team_total_credits)
                    ELSE 0
                END
            ),
            0
        )::float8 AS total_returns
    FROM core.portfolios p
    LEFT JOIN portfolio_investments pi ON pi.portfolio_id = p.id AND pi.pool_id = p.pool_id
    LEFT JOIN core.teams tt ON tt.id = pi.team_id AND tt.deleted_at IS NULL
    WHERE p.deleted_at IS NULL
    GROUP BY p.id, p.pool_id
)
SELECT
    p.id,
    p.name,
    p.user_id,
    p.pool_id,
    p.created_at,
    p.updated_at,
    p.deleted_at,
    COALESCE(pr.total_returns, 0)::float8 AS total_returns
FROM core.portfolios p
LEFT JOIN portfolio_returns pr ON p.id = pr.portfolio_id
WHERE p.pool_id = $1 AND p.deleted_at IS NULL
ORDER BY pr.total_returns DESC NULLS LAST, p.created_at DESC;

-- name: GetPortfolioByID :one
SELECT
    id,
    name,
    user_id,
    pool_id,
    created_at,
    updated_at,
    deleted_at
FROM core.portfolios
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePortfolio :exec
INSERT INTO core.portfolios (id, name, user_id, pool_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());

-- name: SoftDeletePortfolio :execrows
UPDATE core.portfolios
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListDistinctUserIDsByPoolID :many
SELECT DISTINCT user_id
FROM core.portfolios
WHERE pool_id = $1 AND user_id IS NOT NULL AND deleted_at IS NULL;
