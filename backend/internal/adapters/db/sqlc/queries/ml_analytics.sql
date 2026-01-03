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

-- name: ListTournamentSimulationBatchesByCoreTournamentID :many
SELECT
	b.id,
	b.tournament_id,
	b.tournament_state_snapshot_id,
	b.n_sims,
	b.seed,
	b.probability_source_key,
	b.created_at
FROM analytics.tournament_simulation_batches b
WHERE b.tournament_id = $1::uuid
	AND b.deleted_at IS NULL
ORDER BY b.created_at DESC;

-- name: ListCalcuttaEvaluationRunsByCoreCalcuttaID :many
SELECT
	cer.id,
	cer.tournament_simulation_batch_id,
	cer.calcutta_snapshot_id,
	cer.purpose,
	cer.created_at
FROM analytics.calcutta_evaluation_runs cer
JOIN core.calcutta_snapshots cs
	ON cs.id = cer.calcutta_snapshot_id
	AND cs.deleted_at IS NULL
WHERE cs.base_calcutta_id = $1::uuid
	AND cer.deleted_at IS NULL
ORDER BY cer.created_at DESC;

-- name: ListStrategyGenerationRunsByCoreCalcuttaID :many
SELECT
	sgr.id,
	sgr.run_key,
	sgr.tournament_simulation_batch_id,
	sgr.calcutta_id,
	sgr.purpose,
	sgr.returns_model_key,
	sgr.investment_model_key,
	sgr.optimizer_key,
	sgr.params_json,
	sgr.git_sha,
	sgr.created_at
FROM lab_gold.strategy_generation_runs sgr
WHERE sgr.calcutta_id = $1::uuid
	AND sgr.deleted_at IS NULL
ORDER BY sgr.created_at DESC;

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

-- name: GetStrategyGenerationRunByRunKey :one
SELECT
	sgr.id,
	COALESCE(sgr.run_key, ''::text) AS run_id,
	sgr.calcutta_id,
	COALESCE(NULLIF(sgr.optimizer_key::text, ''::text), 'legacy'::text) AS strategy,
	COALESCE(tsb.n_sims, 0)::int AS n_sims,
	COALESCE(tsb.seed, 0)::int AS seed,
	COALESCE(c.budget_points, 100)::int AS budget_points,
	sgr.created_at
FROM lab_gold.strategy_generation_runs sgr
LEFT JOIN core.calcuttas c
	ON c.id = sgr.calcutta_id
	AND c.deleted_at IS NULL
LEFT JOIN analytics.tournament_simulation_batches tsb
	ON tsb.id = sgr.tournament_simulation_batch_id
	AND tsb.deleted_at IS NULL
WHERE sgr.run_key = $1::text
	AND sgr.deleted_at IS NULL
LIMIT 1;

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

-- name: GetOurEntryBidsByStrategyGenerationRunID :many
SELECT
	t.id as team_id,
	t.school_name,
	t.seed,
	t.region,
	reb.recommended_bid_points,
	reb.expected_roi
FROM lab_gold.recommended_entry_bids reb
JOIN lab_bronze.teams t ON t.id = reb.team_id
WHERE reb.strategy_generation_run_id = $1::uuid
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

-- name: GetEntryPortfolioByStrategyGenerationRunID :many
-- For our strategy entry (lineage-native)
SELECT
	t.id as team_id,
	t.school_name,
	t.seed,
	t.region,
	reb.recommended_bid_points as bid_amount
FROM lab_gold.recommended_entry_bids reb
JOIN lab_bronze.teams t ON reb.team_id = t.id
WHERE reb.strategy_generation_run_id = sqlc.arg(strategy_generation_run_id)::uuid
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
JOIN lab_bronze.calcuttas bc ON bc.id = eb.calcutta_id
JOIN lab_gold.strategy_generation_runs sgr
	ON sgr.calcutta_id = bc.core_calcutta_id
	AND sgr.run_key = sqlc.arg(run_id)
	AND sgr.deleted_at IS NULL
WHERE eb.entry_name = sqlc.arg(entry_name)
ORDER BY eb.bid_amount_points DESC;

-- name: GetOptimizationRunsByYear :many
SELECT
	COALESCE(sgr.run_key, ''::text) AS run_id,
	sgr.calcutta_id,
	COALESCE(NULLIF(sgr.optimizer_key::text, ''::text), 'legacy'::text) AS strategy,
	COALESCE(tsb.n_sims, 0)::int AS n_sims,
	COALESCE(tsb.seed, 0)::int AS seed,
	COALESCE(c.budget_points, 100)::int AS budget_points,
	sgr.created_at
FROM lab_gold.strategy_generation_runs sgr
JOIN core.calcuttas c ON c.id = sgr.calcutta_id AND c.deleted_at IS NULL
JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
JOIN core.seasons seas ON seas.id = t.season_id
LEFT JOIN analytics.tournament_simulation_batches tsb
	ON tsb.id = sgr.tournament_simulation_batch_id
	AND tsb.deleted_at IS NULL
WHERE sgr.deleted_at IS NULL
	AND sgr.run_key IS NOT NULL
	AND seas.year = $1::int
ORDER BY sgr.created_at DESC;

-- name: GetLatestStrategyGenerationRunKeyByCoreCalcuttaID :one
WITH srg AS (
	SELECT
		sgr.run_key
	FROM lab_gold.strategy_generation_runs sgr
	WHERE sgr.calcutta_id = $1::uuid
		AND sgr.run_key IS NOT NULL
		AND sgr.deleted_at IS NULL
),
perf AS (
	SELECT
		ep.run_id,
		MAX(ep.created_at) AS created_at
	FROM analytics.entry_performance ep
	JOIN srg ON srg.run_key = ep.run_id
	WHERE ep.deleted_at IS NULL
	GROUP BY ep.run_id
)
SELECT run_id
FROM perf
ORDER BY created_at DESC
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

-- name: GetLatestCalcuttaEvaluationRunIDByCoreCalcuttaID :one
SELECT
	cer.id
FROM analytics.calcutta_evaluation_runs cer
JOIN core.calcutta_snapshots cs
	ON cs.id = cer.calcutta_snapshot_id
	AND cs.deleted_at IS NULL
WHERE cs.base_calcutta_id = $1::uuid
	AND cer.deleted_at IS NULL
ORDER BY cer.created_at DESC
LIMIT 1;

-- name: GetEntryPerformanceByCalcuttaEvaluationRunID :many
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
		WHERE eso.calcutta_evaluation_run_id = $1::uuid
		  AND eso.entry_name = gep.entry_name
		  AND eso.deleted_at IS NULL
	) as total_simulations
FROM analytics.entry_performance gep
WHERE gep.calcutta_evaluation_run_id = $1::uuid
	AND gep.deleted_at IS NULL
ORDER BY gep.mean_payout DESC;

-- name: GetEntryRankingsByRunKey :many
WITH strategy_run AS (
	SELECT
		sgr.id AS strategy_generation_run_id,
		sgr.calcutta_id AS core_calcutta_id
	FROM lab_gold.strategy_generation_runs sgr
	WHERE sgr.run_key = sqlc.arg(run_id)::text
		AND sgr.deleted_at IS NULL
	LIMIT 1
),
lab_calcutta AS (
	SELECT bc.id AS lab_calcutta_id
	FROM lab_bronze.calcuttas bc
	JOIN strategy_run sr ON sr.core_calcutta_id = bc.core_calcutta_id
	LIMIT 1
),
base AS (
	SELECT
		gep.entry_name,
		COALESCE(gep.mean_payout, 0.0)::double precision AS mean_normalized_payout,
		COALESCE(gep.p_top1, 0.0)::double precision AS p_top1,
		COALESCE(gep.p_in_money, 0.0)::double precision AS p_in_money
	FROM analytics.entry_performance gep
	WHERE gep.run_id = sqlc.arg(run_id)::text
		AND gep.deleted_at IS NULL
),
with_totals AS (
	SELECT
		ROW_NUMBER() OVER (ORDER BY b.mean_normalized_payout DESC)::int AS rank,
		b.entry_name,
		(b.entry_name = 'our_strategy')::boolean AS is_our_strategy,
		b.mean_normalized_payout,
		b.p_top1,
		b.p_in_money,
		COUNT(*) OVER ()::int AS total_entries
	FROM base b
),
with_percentile AS (
	SELECT
		wt.*,
		CASE
			WHEN wt.total_entries > 1 THEN (wt.total_entries - wt.rank)::double precision / (wt.total_entries - 1)::double precision
			ELSE 1.0::double precision
		END AS percentile_rank
	FROM with_totals wt
),
with_bids AS (
	SELECT
		wp.rank,
		wp.entry_name,
		wp.is_our_strategy,
		CASE
			WHEN wp.is_our_strategy THEN COALESCE(os.n_teams, 0)
			ELSE COALESCE(ab.n_teams, 0)
		END::int AS n_teams,
		CASE
			WHEN wp.is_our_strategy THEN COALESCE(os.total_bid_points, 0)
			ELSE COALESCE(ab.total_bid_points, 0)
		END::int AS total_bid_points,
		wp.mean_normalized_payout,
		wp.percentile_rank,
		wp.p_top1,
		wp.p_in_money,
		wp.total_entries
	FROM with_percentile wp
	LEFT JOIN (
		SELECT
			COUNT(*)::int AS n_teams,
			COALESCE(SUM(recommended_bid_points), 0)::int AS total_bid_points
		FROM lab_gold.recommended_entry_bids reb
		JOIN strategy_run sr ON sr.strategy_generation_run_id = reb.strategy_generation_run_id
		WHERE reb.deleted_at IS NULL
	) os ON wp.is_our_strategy
	LEFT JOIN (
		SELECT
			eb.entry_name,
			COUNT(*)::int AS n_teams,
			COALESCE(SUM(eb.bid_amount_points), 0)::int AS total_bid_points
		FROM lab_bronze.entry_bids eb
		JOIN lab_calcutta lc ON lc.lab_calcutta_id = eb.calcutta_id
		WHERE eb.deleted_at IS NULL
		GROUP BY eb.entry_name
	) ab ON (NOT wp.is_our_strategy AND ab.entry_name = wp.entry_name)
)
SELECT
	rank,
	entry_name AS entry_key,
	is_our_strategy,
	n_teams,
	total_bid_points,
	mean_normalized_payout,
	percentile_rank,
	p_top1,
	p_in_money,
	total_entries
FROM with_bids
ORDER BY rank ASC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetEntrySimulationsByRunKeyAndEntryName :many
SELECT
	eso.sim_id,
	eso.payout_points,
	eso.points_scored,
	eso.rank,
	(
		SELECT COUNT(*)::int
		FROM analytics.entry_simulation_outcomes eso_all
		WHERE eso_all.run_id = sqlc.arg(run_id)::text
			AND eso_all.sim_id = eso.sim_id
			AND eso_all.deleted_at IS NULL
	) AS n_entries,
	(
		SELECT MAX(eso_all.payout_points)::int
		FROM analytics.entry_simulation_outcomes eso_all
		WHERE eso_all.run_id = sqlc.arg(run_id)::text
			AND eso_all.sim_id = eso.sim_id
			AND eso_all.deleted_at IS NULL
	) AS max_payout_points,
	CASE
		WHEN (
			SELECT MAX(eso_all.payout_points)
			FROM analytics.entry_simulation_outcomes eso_all
			WHERE eso_all.run_id = sqlc.arg(run_id)::text
				AND eso_all.sim_id = eso.sim_id
				AND eso_all.deleted_at IS NULL
		) > 0 THEN (eso.payout_points::double precision / (
			SELECT MAX(eso_all.payout_points)::double precision
			FROM analytics.entry_simulation_outcomes eso_all
			WHERE eso_all.run_id = sqlc.arg(run_id)::text
				AND eso_all.sim_id = eso.sim_id
				AND eso_all.deleted_at IS NULL
		))
		ELSE 0.0::double precision
	END AS normalized_payout,
	(
		SELECT (COUNT(*) > 1)
		FROM analytics.entry_simulation_outcomes eso_tie
		WHERE eso_tie.run_id = sqlc.arg(run_id)::text
			AND eso_tie.sim_id = eso.sim_id
			AND eso_tie.rank = eso.rank
			AND eso_tie.deleted_at IS NULL
	) AS is_tied
FROM analytics.entry_simulation_outcomes eso
WHERE eso.run_id = sqlc.arg(run_id)::text
	AND eso.entry_name = sqlc.arg(entry_name)::text
	AND eso.deleted_at IS NULL
ORDER BY eso.sim_id ASC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetEntrySimulationSummaryByRunKeyAndEntryName :one
WITH per_sim AS (
	SELECT
		sim_id,
		MAX(payout_points)::double precision AS max_payout_points,
		COUNT(*)::int AS n_entries
	FROM analytics.entry_simulation_outcomes
	WHERE run_id = sqlc.arg(run_id)::text
		AND deleted_at IS NULL
	GROUP BY sim_id
),
entry_sims AS (
	SELECT
		eso.sim_id,
		eso.payout_points::double precision AS payout_points,
		eso.points_scored::double precision AS points_scored,
		CASE
			WHEN ps.max_payout_points > 0 THEN (eso.payout_points::double precision / ps.max_payout_points)
			ELSE 0.0::double precision
		END AS normalized_payout
	FROM analytics.entry_simulation_outcomes eso
	JOIN per_sim ps ON ps.sim_id = eso.sim_id
	WHERE eso.run_id = sqlc.arg(run_id)::text
		AND eso.entry_name = sqlc.arg(entry_name)::text
		AND eso.deleted_at IS NULL
)
SELECT
	COUNT(*)::int AS total_simulations,
	COALESCE(AVG(payout_points), 0.0)::double precision AS mean_payout_points,
	COALESCE(AVG(points_scored), 0.0)::double precision AS mean_points,
	COALESCE(AVG(normalized_payout), 0.0)::double precision AS mean_normalized_payout,
	COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY payout_points), 0.0)::double precision AS p50_payout_points,
	COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY payout_points), 0.0)::double precision AS p90_payout_points
FROM entry_sims;

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
