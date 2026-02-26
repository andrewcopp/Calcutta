-- name: ListEntriesByCalcuttaID :many
WITH entry_bids AS (
    SELECT
        cet.entry_id,
        ce.calcutta_id,
        cet.team_id,
        cet.bid_points::float8 AS bid_points,
        SUM(cet.bid_points::float8) OVER (
            PARTITION BY ce.calcutta_id, cet.team_id
        ) AS team_total_bid_points
    FROM core.entry_teams cet
    JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
    WHERE cet.deleted_at IS NULL
),
entry_points AS (
    SELECT
        ce.id AS entry_id,
        COALESCE(
            SUM(
                CASE
                    WHEN eb.team_total_bid_points > 0 THEN
                        core.calcutta_points_for_progress(ce.calcutta_id, tt.wins, tt.byes)::float8
                        * (eb.bid_points / eb.team_total_bid_points)
                    ELSE 0
                END
            ),
            0
        )::float8 AS total_points
    FROM core.entries ce
    LEFT JOIN entry_bids eb ON eb.entry_id = ce.id AND eb.calcutta_id = ce.calcutta_id
    LEFT JOIN core.teams tt ON tt.id = eb.team_id AND tt.deleted_at IS NULL
    WHERE ce.deleted_at IS NULL
    GROUP BY ce.id, ce.calcutta_id
)
SELECT
    ce.id,
    ce.name,
    ce.user_id,
    ce.calcutta_id,
    ce.status,
    ce.created_at,
    ce.updated_at,
    ce.deleted_at,
    COALESCE(ep.total_points, 0)::float8 AS total_points
FROM core.entries ce
LEFT JOIN entry_points ep ON ce.id = ep.entry_id
WHERE ce.calcutta_id = $1 AND ce.deleted_at IS NULL
ORDER BY ep.total_points DESC NULLS LAST, ce.created_at DESC;

-- name: GetEntryByID :one
SELECT
    id,
    name,
    user_id,
    calcutta_id,
    status,
    created_at,
    updated_at,
    deleted_at
FROM core.entries
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateEntry :exec
INSERT INTO core.entries (id, name, user_id, calcutta_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());

-- name: UpdateEntryStatus :exec
UPDATE core.entries
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListDistinctUserIDsByCalcuttaID :many
SELECT DISTINCT user_id
FROM core.entries
WHERE calcutta_id = $1 AND user_id IS NOT NULL AND deleted_at IS NULL;
