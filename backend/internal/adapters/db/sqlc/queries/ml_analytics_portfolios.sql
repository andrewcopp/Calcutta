-- name: GetOurEntryBidsByRunID :many
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    reb.bid_points,
    reb.expected_roi
FROM derived.recommended_entry_bids reb
JOIN derived.teams t ON t.id = reb.team_id
WHERE reb.run_id = $1::text
ORDER BY reb.bid_points DESC;

-- name: GetOurEntryBidsByStrategyGenerationRunID :many
SELECT
	t.id as team_id,
	t.school_name,
	t.seed,
	t.region,
	reb.bid_points,
	reb.expected_roi
FROM derived.recommended_entry_bids reb
JOIN derived.teams t ON t.id = reb.team_id
WHERE reb.strategy_generation_run_id = $1::uuid
ORDER BY reb.bid_points DESC;

-- name: GetEntryPortfolio :many
-- For our strategy entry
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    reb.bid_points as bid_points
FROM derived.recommended_entry_bids reb
JOIN derived.teams t ON reb.team_id = t.id
WHERE reb.run_id = sqlc.arg(run_id)
ORDER BY reb.bid_points DESC;

-- name: GetEntryPortfolioByStrategyGenerationRunID :many
-- For our strategy entry (lineage-native)
SELECT
	t.id as team_id,
	t.school_name,
	t.seed,
	t.region,
	reb.bid_points as bid_points
FROM derived.recommended_entry_bids reb
JOIN derived.teams t ON reb.team_id = t.id
WHERE reb.strategy_generation_run_id = sqlc.arg(strategy_generation_run_id)::uuid
ORDER BY reb.bid_points DESC;

-- name: GetActualEntryPortfolio :many
-- For actual entries from the auction
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    eb.bid_points
FROM derived.entry_bids eb
JOIN derived.teams t ON eb.team_id = t.id
JOIN derived.calcuttas bc ON bc.id = eb.calcutta_id
JOIN derived.strategy_generation_runs sgr
	ON sgr.calcutta_id = bc.core_calcutta_id
	AND sgr.run_key = sqlc.arg(run_id)
	AND sgr.deleted_at IS NULL
WHERE eb.entry_name = sqlc.arg(entry_name)
ORDER BY eb.bid_points DESC;
