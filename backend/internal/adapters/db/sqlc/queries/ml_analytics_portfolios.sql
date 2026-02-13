-- name: GetOurEntryBidsByRunID :many
SELECT
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    (bid->>'bid_points')::int as bid_points,
    COALESCE((bid->>'expected_roi')::double precision, 0) as expected_roi
FROM derived.optimized_entries oe,
     jsonb_array_elements(oe.bids_json) AS bid
JOIN core.teams t ON t.id = (bid->>'team_id')::uuid AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE oe.run_key = $1::text
    AND oe.deleted_at IS NULL
ORDER BY (bid->>'bid_points')::int DESC;

-- name: GetOurEntryBidsByOptimizedEntryID :many
SELECT
	t.id as team_id,
	s.name as school_name,
	t.seed,
	t.region,
	(bid->>'bid_points')::int as bid_points,
	COALESCE((bid->>'expected_roi')::double precision, 0) as expected_roi
FROM derived.optimized_entries oe,
     jsonb_array_elements(oe.bids_json) AS bid
JOIN core.teams t ON t.id = (bid->>'team_id')::uuid AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE oe.id = $1::uuid
    AND oe.deleted_at IS NULL
ORDER BY (bid->>'bid_points')::int DESC;

-- name: GetEntryPortfolio :many
-- For our strategy entry
SELECT
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    (bid->>'bid_points')::int as bid_points
FROM derived.optimized_entries oe,
     jsonb_array_elements(oe.bids_json) AS bid
JOIN core.teams t ON (bid->>'team_id')::uuid = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE oe.run_key = sqlc.arg(run_id)
    AND oe.deleted_at IS NULL
ORDER BY (bid->>'bid_points')::int DESC;

-- name: GetEntryPortfolioByOptimizedEntryID :many
-- For our strategy entry (lineage-native)
SELECT
	t.id as team_id,
	s.name as school_name,
	t.seed,
	t.region,
	(bid->>'bid_points')::int as bid_points
FROM derived.optimized_entries oe,
     jsonb_array_elements(oe.bids_json) AS bid
JOIN core.teams t ON (bid->>'team_id')::uuid = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE oe.id = sqlc.arg(optimized_entry_id)::uuid
    AND oe.deleted_at IS NULL
ORDER BY (bid->>'bid_points')::int DESC;

-- name: GetActualEntryPortfolio :many
-- For actual entries from the auction
WITH optimized_entry AS (
	SELECT oe.calcutta_id
	FROM derived.optimized_entries oe
	WHERE oe.run_key = sqlc.arg(run_id)::text
		AND oe.deleted_at IS NULL
	LIMIT 1
)
SELECT
    t.id as team_id,
    s.name as school_name,
    t.seed,
    t.region,
    eb.bid_points
FROM derived.entry_bids eb
JOIN optimized_entry oe ON oe.calcutta_id = eb.calcutta_id
JOIN core.teams t ON eb.team_id = t.id AND t.deleted_at IS NULL
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
WHERE eb.entry_name = sqlc.arg(entry_name)
    AND eb.deleted_at IS NULL
ORDER BY eb.bid_points DESC;
