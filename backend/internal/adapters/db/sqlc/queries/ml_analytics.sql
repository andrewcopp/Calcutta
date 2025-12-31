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
FROM bronze_tournaments t
JOIN silver_simulated_tournaments st ON t.id = st.tournament_id
WHERE t.season = $1::int
GROUP BY t.id, t.season;

-- name: GetTeamPerformanceByID :one
WITH round_distribution AS (
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
    FROM silver_simulated_tournaments st
    JOIN bronze_teams t ON t.id = st.team_id
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
    AVG(CASE (st.wins + st.byes)
        WHEN 0 THEN 0
        WHEN 1 THEN 0
        WHEN 2 THEN 50
        WHEN 3 THEN 150
        WHEN 4 THEN 300
        WHEN 5 THEN 500
        WHEN 6 THEN 750
        WHEN 7 THEN 1050
        ELSE 0
    END)::float as avg_points,
    jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM bronze_teams t
JOIN silver_simulated_tournaments st ON st.team_id = t.id
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
FROM bronze_teams t
JOIN bronze_tournaments bt ON bt.id = t.tournament_id
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
FROM gold_optimization_runs
WHERE run_id = $1::text;

-- name: GetOurEntryBidsByRunID :many
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    reb.recommended_bid_points,
    reb.expected_roi
FROM gold_recommended_entry_bids reb
JOIN bronze_teams t ON t.id = reb.team_id
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
FROM gold_recommended_entry_bids reb
JOIN bronze_teams t ON reb.team_id = t.id
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
FROM bronze_entry_bids eb
JOIN bronze_teams t ON eb.team_id = t.id
JOIN gold_optimization_runs r ON eb.calcutta_id = r.calcutta_id
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
FROM gold_optimization_runs r
JOIN bronze_calcuttas bc ON bc.id = r.calcutta_id
JOIN bronze_tournaments bt ON bt.id = bc.tournament_id
WHERE bt.season = $1::int
ORDER BY r.created_at DESC;
