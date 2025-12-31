-- ML Analytics Queries
-- For tournament simulation and entry evaluation data

-- name: GetTournamentSimStatsByYear :one
SELECT 
    t.tournament_key,
    t.season,
    COUNT(DISTINCT st.sim_id)::int as n_sims,
    COUNT(DISTINCT st.team_key)::int as n_teams,
    AVG(st.wins + st.byes)::float as avg_progress,
    MAX(st.wins + st.byes)::int as max_progress
FROM bronze_tournaments t
JOIN bronze_simulated_tournaments st ON t.tournament_key = st.tournament_key
WHERE t.season = $1::int
GROUP BY t.tournament_key, t.season;

-- name: GetTeamPerformanceByKey :one
WITH round_distribution AS (
    SELECT 
        st.team_key,
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
    FROM bronze_simulated_tournaments st
    JOIN bronze_teams t ON t.team_key = st.team_key
    WHERE st.team_key = $1::text
    GROUP BY st.team_key, round_name
)
SELECT 
    t.team_key,
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
    tv.p_champion,
    tv.p_finals,
    tv.p_final_four,
    tv.p_elite_eight,
    tv.p_sweet_sixteen,
    tv.p_round_32,
    jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM bronze_teams t
JOIN bronze_simulated_tournaments st ON st.team_key = t.team_key
LEFT JOIN silver_team_tournament_value tv ON tv.team_key = t.team_key
LEFT JOIN round_distribution rd ON rd.team_key = t.team_key
WHERE t.team_key = $1::text
GROUP BY t.team_key, t.school_name, t.seed, t.region, t.kenpom_net,
         tv.p_champion, tv.p_finals, tv.p_final_four, tv.p_elite_eight,
         tv.p_sweet_sixteen, tv.p_round_32;

-- name: GetTeamPredictionsByYear :many
SELECT 
    t.team_key,
    t.school_name,
    t.seed,
    t.region,
    tv.expected_points,
    pms.predicted_share_of_pool as predicted_market_share,
    (pms.predicted_share_of_pool * bc.budget_points)::float as predicted_market_points,
    tv.p_champion,
    t.kenpom_net
FROM bronze_teams t
JOIN bronze_tournaments bt ON bt.tournament_key = t.tournament_key
JOIN bronze_calcuttas bc ON bc.tournament_key = bt.tournament_key
LEFT JOIN silver_team_tournament_value tv ON tv.team_key = t.team_key
LEFT JOIN silver_predicted_market_share pms ON pms.team_key = t.team_key AND pms.calcutta_key = bc.calcutta_key
WHERE bt.season = $1::int
ORDER BY tv.expected_points DESC NULLS LAST;

-- name: GetOptimizationRunByID :one
SELECT 
    run_id,
    calcutta_key,
    strategy,
    n_sims,
    seed,
    budget_points,
    run_timestamp
FROM gold_optimization_runs
WHERE run_id = $1::text;

-- name: GetOurEntryBidsByRunID :many
SELECT 
    t.team_key,
    t.school_name,
    t.seed,
    t.region,
    reb.bid_amount_points,
    dir.expected_points,
    dir.predicted_market_points,
    dir.actual_market_points,
    dir.our_ownership,
    dir.expected_roi,
    dir.our_roi,
    dir.roi_degradation
FROM gold_recommended_entry_bids reb
JOIN bronze_teams t ON t.team_key = reb.team_key
LEFT JOIN gold_detailed_investment_report dir ON dir.run_id = reb.run_id AND dir.team_key = reb.team_key
WHERE reb.run_id = $1::text
ORDER BY reb.bid_amount_points DESC;

-- name: GetEntryPerformanceByRunID :one
SELECT 
    ep.n_teams,
    ep.total_bid_points,
    ep.mean_normalized_payout,
    ep.p_top1,
    ep.p_in_money,
    ep.percentile_rank
FROM gold_entry_performance ep
WHERE ep.run_id = $1::text AND ep.is_our_strategy = true;

-- name: GetEntryRankingsByRunID :many
SELECT 
    er.rank,
    er.entry_key,
    er.is_our_strategy,
    ep.n_teams,
    ep.total_bid_points,
    ep.mean_normalized_payout,
    ep.percentile_rank,
    ep.p_top1,
    ep.p_in_money,
    er.total_entries
FROM view_entry_rankings er
JOIN gold_entry_performance ep ON ep.run_id = er.run_id AND ep.entry_key = er.entry_key
WHERE er.run_id = $1::text
ORDER BY er.rank
LIMIT $2::int OFFSET $3::int;

-- name: GetEntrySimulationsByKey :many
SELECT 
    sim_id,
    payout_cents,
    total_points,
    finish_position,
    is_tied,
    normalized_payout,
    n_entries
FROM gold_entry_simulation_outcomes
WHERE run_id = $1::text AND entry_key = $2::text
ORDER BY payout_cents DESC, total_points DESC
LIMIT $3::int OFFSET $4::int;

-- name: GetEntrySimulationSummary :one
SELECT 
    COUNT(*)::int as total_simulations,
    AVG(payout_cents)::float as mean_payout_cents,
    AVG(total_points)::float as mean_points,
    AVG(normalized_payout)::float as mean_normalized_payout,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY payout_cents)::int as p50_payout_cents,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY payout_cents)::int as p90_payout_cents
FROM gold_entry_simulation_outcomes
WHERE run_id = $1::text AND entry_key = $2::text;

-- name: GetEntryPortfolio :many
-- For our strategy entry
SELECT 
    t.team_key,
    t.school_name,
    t.seed,
    t.region,
    reb.bid_amount_points as bid_amount
FROM gold_recommended_entry_bids reb
JOIN bronze_teams t ON reb.team_key = t.team_key
WHERE reb.run_id = sqlc.arg(run_id)
ORDER BY reb.bid_amount_points DESC;

-- name: GetActualEntryPortfolio :many
-- For actual entries from the auction
SELECT 
    t.team_key,
    t.school_name,
    t.seed,
    t.region,
    eb.bid_amount
FROM bronze_entry_bids eb
JOIN bronze_teams t ON eb.team_key = t.team_key
JOIN gold_optimization_runs r ON eb.calcutta_key = r.calcutta_key
WHERE r.run_id = sqlc.arg(run_id) AND eb.entry_key = sqlc.arg(entry_key)
ORDER BY eb.bid_amount DESC;

-- name: GetOptimizationRunsByYear :many
SELECT 
    r.run_id,
    r.calcutta_key,
    r.strategy,
    r.n_sims,
    r.seed,
    r.budget_points,
    r.run_timestamp
FROM gold_optimization_runs r
JOIN bronze_calcuttas bc ON bc.calcutta_key = r.calcutta_key
JOIN bronze_tournaments bt ON bt.tournament_key = bc.tournament_key
WHERE bt.season = $1::int
ORDER BY r.run_timestamp DESC;
