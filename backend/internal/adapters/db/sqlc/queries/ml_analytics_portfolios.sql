-- name: GetOurEntryBidsByRunID :many
SELECT 
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    reb.bid_points,
    reb.expected_roi
FROM derived.recommended_entry_bids reb
JOIN core.teams t ON t.id = reb.team_id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE reb.run_id = $1::text
    AND reb.deleted_at IS NULL
ORDER BY reb.bid_points DESC;

-- name: GetOurEntryBidsByStrategyGenerationRunID :many
SELECT
	t.id as team_id,
	s.name as school_name,
	t.seed,
	t.region,
	reb.bid_points,
	reb.expected_roi
FROM derived.recommended_entry_bids reb
JOIN core.teams t ON t.id = reb.team_id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE reb.strategy_generation_run_id = $1::uuid
    AND reb.deleted_at IS NULL
ORDER BY reb.bid_points DESC;

-- name: GetEntryPortfolio :many
-- For our strategy entry
SELECT 
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    reb.bid_points as bid_points
FROM derived.recommended_entry_bids reb
JOIN core.teams t ON reb.team_id = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE reb.run_id = sqlc.arg(run_id)
    AND reb.deleted_at IS NULL
ORDER BY reb.bid_points DESC;

-- name: GetEntryPortfolioByStrategyGenerationRunID :many
-- For our strategy entry (lineage-native)
SELECT
	t.id as team_id,
	s.name as school_name,
	t.seed,
	t.region,
	reb.bid_points as bid_points
FROM derived.recommended_entry_bids reb
JOIN core.teams t ON reb.team_id = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE reb.strategy_generation_run_id = sqlc.arg(strategy_generation_run_id)::uuid
    AND reb.deleted_at IS NULL
ORDER BY reb.bid_points DESC;

-- name: GetActualEntryPortfolio :many
-- For actual entries from the auction
WITH strategy_run AS (
	SELECT sgr.calcutta_id
	FROM derived.strategy_generation_runs sgr
	WHERE sgr.run_key = sqlc.arg(run_id)::text
		AND sgr.deleted_at IS NULL
	LIMIT 1
)
SELECT 
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    eb.bid_points
FROM derived.entry_bids eb
JOIN strategy_run sr ON sr.calcutta_id = eb.calcutta_id
JOIN core.teams t ON eb.team_id = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE eb.entry_name = sqlc.arg(entry_name)
    AND eb.deleted_at IS NULL
ORDER BY eb.bid_points DESC;
