-- ML Analytics Queries
-- For tournament simulation and entry evaluation data

-- name: GetTournamentSimStatsByYear :one
SELECT 
    t.id as tournament_id,
    t.season,
    COUNT(DISTINCT st.sim_id)::int as n_sims,
    COUNT(DISTINCT st.team_id)::int as n_teams,
    AVG(st.wins + st.byes)::float as avg_progress,
    MAX(st.wins + st.byes)::int as max_progress
FROM lab_bronze.tournaments t
JOIN analytics.simulated_tournaments st ON t.id = st.tournament_id
WHERE t.season = $1::int
GROUP BY t.id, t.season;

-- name: GetTournamentSimStatsByCoreTournamentID :one
WITH tournament_info AS (
	SELECT
		bt.id as tournament_id,
		bt.season
	FROM lab_bronze.tournaments bt
	WHERE bt.core_tournament_id = sqlc.arg(core_tournament_id)::uuid
	LIMIT 1
),
sim_stats AS (
	SELECT
		COUNT(DISTINCT sim_id)::int as total_simulations,
		COUNT(DISTINCT team_id)::int as total_teams
	FROM analytics.simulated_tournaments st
	JOIN tournament_info ti ON st.tournament_id = ti.tournament_id
),
prediction_stats AS (
	SELECT COUNT(*)::int as total_predictions
	FROM lab_silver.predicted_game_outcomes pgo
	JOIN tournament_info ti ON pgo.tournament_id = ti.tournament_id
),
win_stats AS (
	SELECT
		AVG(wins)::double precision as mean_wins,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY wins)::double precision as median_wins,
		MAX(wins)::int as max_wins
	FROM analytics.simulated_tournaments st
	JOIN tournament_info ti ON st.tournament_id = ti.tournament_id
)
SELECT
	ti.tournament_id,
	ti.season,
	COALESCE(ss.total_simulations, 0)::int as total_simulations,
	COALESCE(ps.total_predictions, 0)::int as total_predictions,
	COALESCE(ws.mean_wins, 0.0)::double precision as mean_wins,
	COALESCE(ws.median_wins, 0.0)::double precision as median_wins,
	COALESCE(ws.max_wins, 0)::int as max_wins,
	NOW()::timestamptz as last_updated
FROM tournament_info ti
LEFT JOIN sim_stats ss ON true
LEFT JOIN prediction_stats ps ON true
LEFT JOIN win_stats ws ON true;

-- name: GetTeamPerformanceByID :one
WITH season_ctx AS (
    SELECT bt.core_tournament_id
    FROM lab_bronze.teams t
    JOIN lab_bronze.tournaments bt ON bt.id = t.tournament_id
    WHERE t.id = $1::uuid
    LIMIT 1
),
main_tournament AS (
    SELECT tr.id
    FROM core.tournaments tr
    JOIN season_ctx sc ON tr.id = sc.core_tournament_id
    WHERE tr.deleted_at IS NULL
    ORDER BY tr.created_at DESC
    LIMIT 1
),
calcutta_ctx AS (
    SELECT c.id AS calcutta_id
    FROM core.calcuttas c
    JOIN main_tournament mt ON mt.id = c.tournament_id
    WHERE c.deleted_at IS NULL
    ORDER BY c.created_at DESC
    LIMIT 1
),
round_distribution AS (
    SELECT 
        st.team_id,
        CASE (st.wins + st.byes)
            WHEN 0 THEN 'R64'
            WHEN 1 THEN 'R64'
            WHEN 2 THEN 'R32'
            WHEN 3 THEN 'S16'
            WHEN 4 THEN 'E8'
            WHEN 5 THEN 'F4'
            WHEN 6 THEN 'Finals'
            WHEN 7 THEN 'Champion'
            ELSE 'Unknown'
        END as round_name,
        COUNT(*)::int as count
    FROM analytics.simulated_tournaments st
    JOIN lab_bronze.teams t ON t.id = st.team_id
    WHERE st.team_id = $1::uuid
    GROUP BY st.team_id, round_name
)
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    t.kenpom_net,
    COUNT(DISTINCT st.sim_id)::int as total_sims,
    AVG(st.wins)::float as avg_wins,
    AVG(
        CASE
            WHEN (SELECT calcutta_id FROM calcutta_ctx) IS NULL THEN 0
            ELSE core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins, st.byes)
        END
    )::float as avg_points,
    jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM lab_bronze.teams t
JOIN analytics.simulated_tournaments st ON st.team_id = t.id
LEFT JOIN round_distribution rd ON rd.team_id = t.id
WHERE t.id = $1::uuid
GROUP BY t.id, t.school_name, t.seed, t.region, t.kenpom_net;

-- name: GetTeamPredictionsByYear :many
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    t.kenpom_net
FROM lab_bronze.teams t
JOIN lab_bronze.tournaments bt ON bt.id = t.tournament_id
WHERE bt.season = $1::int
ORDER BY t.seed;

-- name: GetOptimizationRunByID :one
SELECT 
    run_id,
    calcutta_id,
    strategy,
    n_sims,
    seed,
    budget_points,
    created_at
FROM lab_gold.optimization_runs
WHERE run_id = $1::text;

-- name: GetOurEntryBidsByRunID :many
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    reb.recommended_bid_points,
    reb.expected_roi
FROM lab_gold.recommended_entry_bids reb
JOIN lab_bronze.teams t ON t.id = reb.team_id
WHERE reb.run_id = $1::text
ORDER BY reb.recommended_bid_points DESC;

-- Removed - table doesn't exist in new schema

-- Removed - view doesn't exist in new schema

-- Removed - table schema changed

-- Removed - table schema changed

-- name: GetEntryPortfolio :many
-- For our strategy entry
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    reb.recommended_bid_points as bid_amount
FROM lab_gold.recommended_entry_bids reb
JOIN lab_bronze.teams t ON reb.team_id = t.id
WHERE reb.run_id = sqlc.arg(run_id)
ORDER BY reb.recommended_bid_points DESC;

-- name: GetActualEntryPortfolio :many
-- For actual entries from the auction
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    eb.bid_amount_points
FROM lab_bronze.entry_bids eb
JOIN lab_bronze.teams t ON eb.team_id = t.id
JOIN lab_gold.optimization_runs r ON eb.calcutta_id = r.calcutta_id
WHERE r.run_id = sqlc.arg(run_id) AND eb.entry_name = sqlc.arg(entry_name)
ORDER BY eb.bid_amount_points DESC;

-- name: GetOptimizationRunsByYear :many
SELECT 
    r.run_id,
    r.calcutta_id,
    r.strategy,
    r.n_sims,
    r.seed,
    r.budget_points,
    r.created_at
FROM lab_gold.optimization_runs r
JOIN lab_bronze.calcuttas bc ON bc.id = r.calcutta_id
JOIN lab_bronze.tournaments bt ON bt.id = bc.tournament_id
WHERE bt.season = $1::int
ORDER BY r.created_at DESC;

-- name: GetLatestOptimizationRunIDByCoreCalcuttaID :one
SELECT gor.run_id
FROM lab_gold.optimization_runs gor
JOIN lab_bronze.calcuttas bc ON bc.id = gor.calcutta_id
WHERE bc.core_calcutta_id = $1::uuid
ORDER BY gor.created_at DESC
LIMIT 1;

-- name: GetEntryPerformanceByRunID :many
SELECT
    ROW_NUMBER() OVER (ORDER BY gep.mean_payout DESC)::int as rank,
    gep.entry_name,
    COALESCE(gep.mean_payout, 0.0)::double precision as mean_payout,
    COALESCE(gep.median_payout, 0.0)::double precision as median_payout,
    COALESCE(gep.p_top1, 0.0)::double precision as p_top1,
    COALESCE(gep.p_in_money, 0.0)::double precision as p_in_money,
    (
        SELECT COUNT(*)::int
        FROM analytics.entry_simulation_outcomes eso
        WHERE eso.run_id = $1::text AND eso.entry_name = gep.entry_name
    ) as total_simulations
FROM analytics.entry_performance gep
WHERE gep.run_id = $1::text
ORDER BY gep.mean_payout DESC;

-- name: GetTeamPerformanceByCalcutta :one
WITH calcutta AS (
	SELECT
		c.id AS calcutta_id,
		c.tournament_id
	FROM core.calcuttas c
	WHERE c.id = sqlc.arg(calcutta_id)::uuid
	  AND c.deleted_at IS NULL
	LIMIT 1
),
team_ctx AS (
	SELECT
		t.id AS team_id,
		bt.core_tournament_id
	FROM lab_bronze.teams t
	JOIN lab_bronze.tournaments bt ON bt.id = t.tournament_id
	WHERE t.id = sqlc.arg(team_id)::uuid
	LIMIT 1
),
valid AS (
	SELECT 1
	FROM calcutta c
	JOIN team_ctx tc ON tc.core_tournament_id = c.tournament_id
	LIMIT 1
),
round_distribution AS (
	SELECT
		st.team_id,
		CASE (st.wins + st.byes)
			WHEN 0 THEN 'R64'
			WHEN 1 THEN 'R64'
			WHEN 2 THEN 'R32'
			WHEN 3 THEN 'S16'
			WHEN 4 THEN 'E8'
			WHEN 5 THEN 'F4'
			WHEN 6 THEN 'Finals'
			WHEN 7 THEN 'Champion'
			ELSE 'Unknown'
		END as round_name,
		COUNT(*)::int as count
	FROM analytics.simulated_tournaments st
	WHERE st.team_id = sqlc.arg(team_id)::uuid
	GROUP BY st.team_id, round_name
)
SELECT
	t.id as team_id,
	t.school_name,
	t.seed,
	t.region,
	t.kenpom_net,
	COUNT(DISTINCT st.sim_id)::int as total_sims,
	AVG(st.wins)::float as avg_wins,
	AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float as avg_points,
	jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM lab_bronze.teams t
JOIN valid v ON true

JOIN analytics.simulated_tournaments st ON st.team_id = t.id
	LEFT JOIN round_distribution rd ON rd.team_id = t.id
WHERE t.id = sqlc.arg(team_id)::uuid
GROUP BY t.id, t.school_name, t.seed, t.region, t.kenpom_net;
