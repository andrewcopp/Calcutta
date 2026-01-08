-- name: GetOptimizationRunsByYear :many
SELECT
	COALESCE(sgr.run_key, ''::text) AS run_id,
	COALESCE(NULLIF(sgr.name, ''::text), COALESCE(sgr.run_key, ''::text)) AS name,
	sgr.calcutta_id,
	COALESCE(NULLIF(sgr.optimizer_key::text, ''::text), 'legacy'::text) AS strategy,
	COALESCE(tsb.n_sims, 0)::int AS n_sims,
	COALESCE(tsb.seed, 0)::int AS seed,
	COALESCE(c.budget_points, 100)::int AS budget_points,
	sgr.created_at
FROM derived.strategy_generation_runs sgr
JOIN core.calcuttas c ON c.id = sgr.calcutta_id AND c.deleted_at IS NULL
JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
JOIN core.seasons seas ON seas.id = t.season_id
LEFT JOIN derived.simulated_tournaments tsb
	ON tsb.id = sgr.simulated_tournament_id
	AND tsb.deleted_at IS NULL
WHERE sgr.deleted_at IS NULL
	AND sgr.run_key IS NOT NULL
	AND seas.year = $1::int
ORDER BY sgr.created_at DESC;

-- name: GetLatestStrategyGenerationRunKeyByCoreCalcuttaID :one
WITH srg AS (
	SELECT
		sgr.run_key
	FROM derived.strategy_generation_runs sgr
	WHERE sgr.calcutta_id = $1::uuid
		AND sgr.run_key IS NOT NULL
		AND sgr.deleted_at IS NULL
),
perf AS (
	SELECT
		ep.run_id,
		MAX(ep.created_at) AS created_at
	FROM derived.entry_performance ep
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
    ROW_NUMBER() OVER (ORDER BY gep.mean_normalized_payout DESC)::int as rank,
    gep.entry_name,
    COALESCE(gep.mean_normalized_payout, 0.0)::double precision as mean_normalized_payout,
    COALESCE(gep.median_normalized_payout, 0.0)::double precision as median_normalized_payout,
    COALESCE(gep.p_top1, 0.0)::double precision as p_top1,
    COALESCE(gep.p_in_money, 0.0)::double precision as p_in_money,
    COALESCE((
		SELECT st.n_sims::int
		FROM derived.calcutta_evaluation_runs cer
		JOIN derived.simulated_tournaments st
			ON st.id = cer.simulated_tournament_id
			AND st.deleted_at IS NULL
		WHERE cer.id = gep.calcutta_evaluation_run_id
			AND cer.deleted_at IS NULL
		LIMIT 1
	), 0)::int as total_simulations
FROM derived.entry_performance gep
WHERE gep.run_id = $1::text
ORDER BY gep.mean_normalized_payout DESC;

-- name: GetLatestCalcuttaEvaluationRunIDByCoreCalcuttaID :one
SELECT
	cer.id
FROM derived.calcutta_evaluation_runs cer
JOIN core.calcutta_snapshots cs
	ON cs.id = cer.calcutta_snapshot_id
	AND cs.deleted_at IS NULL
WHERE cs.base_calcutta_id = $1::uuid
	AND cer.deleted_at IS NULL
ORDER BY cer.created_at DESC
LIMIT 1;

-- name: GetEntryPerformanceByCalcuttaEvaluationRunID :many
SELECT
	ROW_NUMBER() OVER (ORDER BY gep.mean_normalized_payout DESC)::int as rank,
	gep.entry_name,
	COALESCE(gep.mean_normalized_payout, 0.0)::double precision as mean_normalized_payout,
	COALESCE(gep.median_normalized_payout, 0.0)::double precision as median_normalized_payout,
	COALESCE(gep.p_top1, 0.0)::double precision as p_top1,
	COALESCE(gep.p_in_money, 0.0)::double precision as p_in_money,
	COALESCE((
		SELECT st.n_sims::int
		FROM derived.calcutta_evaluation_runs cer
		JOIN derived.simulated_tournaments st
			ON st.id = cer.simulated_tournament_id
			AND st.deleted_at IS NULL
		WHERE cer.id = $1::uuid
			AND cer.deleted_at IS NULL
		LIMIT 1
	), 0)::int as total_simulations
FROM derived.entry_performance gep
WHERE gep.calcutta_evaluation_run_id = $1::uuid
	AND gep.deleted_at IS NULL
ORDER BY gep.mean_normalized_payout DESC;

-- name: GetEntryRankingsByRunKey :many
WITH strategy_run AS (
	SELECT
		sgr.id AS strategy_generation_run_id,
		sgr.calcutta_id AS core_calcutta_id
	FROM derived.strategy_generation_runs sgr
	WHERE sgr.run_key = sqlc.arg(run_id)::text
		AND sgr.deleted_at IS NULL
	LIMIT 1
),
focus AS (
	SELECT se.display_name
	FROM derived.simulation_runs sr
	JOIN core.calcutta_snapshot_entries se
		ON se.id = sr.focus_snapshot_entry_id
		AND se.deleted_at IS NULL
	JOIN strategy_run r ON r.strategy_generation_run_id = sr.strategy_generation_run_id
	WHERE sr.deleted_at IS NULL
		AND sr.focus_snapshot_entry_id IS NOT NULL
	ORDER BY sr.created_at DESC
	LIMIT 1
),
base AS (
	SELECT
		gep.entry_name,
		COALESCE(gep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
		COALESCE(gep.p_top1, 0.0)::double precision AS p_top1,
		COALESCE(gep.p_in_money, 0.0)::double precision AS p_in_money
	FROM derived.entry_performance gep
	WHERE gep.run_id = sqlc.arg(run_id)::text
		AND gep.deleted_at IS NULL
),
with_totals AS (
	SELECT
		ROW_NUMBER() OVER (ORDER BY b.mean_normalized_payout DESC)::int AS rank,
		b.entry_name,
		COALESCE((b.entry_name = (SELECT display_name FROM focus)), false)::boolean AS is_our_strategy,
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
			COALESCE(SUM(bid_points), 0)::int AS total_bid_points
		FROM derived.recommended_entry_bids reb
		JOIN strategy_run sr ON sr.strategy_generation_run_id = reb.strategy_generation_run_id
		WHERE reb.deleted_at IS NULL
	) os ON wp.is_our_strategy
	LEFT JOIN (
		SELECT
			eb.entry_name,
			COUNT(*)::int AS n_teams,
			COALESCE(SUM(eb.bid_points), 0)::int AS total_bid_points
		FROM derived.entry_bids eb
		JOIN strategy_run sr ON sr.core_calcutta_id = eb.calcutta_id
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

-- name: GetOurEntryPerformanceSummaryByRunKey :one
WITH base AS (
	SELECT
		gep.entry_name,
		COALESCE(gep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
		COALESCE(gep.p_top1, 0.0)::double precision AS p_top1,
		COALESCE(gep.p_in_money, 0.0)::double precision AS p_in_money
	FROM derived.entry_performance gep
	WHERE gep.run_id = sqlc.arg(run_id)::text
		AND gep.deleted_at IS NULL
),
strategy_run AS (
	SELECT sgr.id AS strategy_generation_run_id
	FROM derived.strategy_generation_runs sgr
	WHERE sgr.run_key = sqlc.arg(run_id)::text
		AND sgr.deleted_at IS NULL
	LIMIT 1
),
focus AS (
	SELECT se.display_name
	FROM derived.simulation_runs sr
	JOIN core.calcutta_snapshot_entries se
		ON se.id = sr.focus_snapshot_entry_id
		AND se.deleted_at IS NULL
	JOIN strategy_run r ON r.strategy_generation_run_id = sr.strategy_generation_run_id
	WHERE sr.deleted_at IS NULL
		AND sr.focus_snapshot_entry_id IS NOT NULL
	ORDER BY sr.created_at DESC
	LIMIT 1
),
with_totals AS (
	SELECT
		ROW_NUMBER() OVER (ORDER BY b.mean_normalized_payout DESC)::int AS rank,
		b.entry_name,
		COALESCE((b.entry_name = (SELECT display_name FROM focus)), false)::boolean AS is_our_strategy,
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
)
SELECT
	mean_normalized_payout,
	p_top1,
	p_in_money,
	percentile_rank
FROM with_percentile
WHERE is_our_strategy
ORDER BY rank ASC
LIMIT 1;

-- name: GetEntrySimulationsByRunKeyAndEntryName :many
SELECT
	eso.sim_id,
	eso.payout_cents,
	eso.points_scored,
	eso.rank,
	(
		SELECT COUNT(*)::int
		FROM derived.entry_simulation_outcomes eso_all
		WHERE eso_all.run_id = sqlc.arg(run_id)::text
			AND eso_all.sim_id = eso.sim_id
			AND eso_all.deleted_at IS NULL
	) AS n_entries,
	(
		SELECT MAX(eso_all.payout_cents)::int
		FROM derived.entry_simulation_outcomes eso_all
		WHERE eso_all.run_id = sqlc.arg(run_id)::text
			AND eso_all.sim_id = eso.sim_id
			AND eso_all.deleted_at IS NULL
	) AS max_payout_cents,
	CASE
		WHEN (
			SELECT MAX(eso_all.payout_cents)
			FROM derived.entry_simulation_outcomes eso_all
			WHERE eso_all.run_id = sqlc.arg(run_id)::text
				AND eso_all.sim_id = eso.sim_id
				AND eso_all.deleted_at IS NULL
		) > 0 THEN (eso.payout_cents::double precision / (
			SELECT MAX(eso_all.payout_cents)::double precision
			FROM derived.entry_simulation_outcomes eso_all
			WHERE eso_all.run_id = sqlc.arg(run_id)::text
				AND eso_all.sim_id = eso.sim_id
				AND eso_all.deleted_at IS NULL
		))
		ELSE 0.0::double precision
	END AS normalized_payout,
	(
		SELECT (COUNT(*) > 1)
		FROM derived.entry_simulation_outcomes eso_tie
		WHERE eso_tie.run_id = sqlc.arg(run_id)::text
			AND eso_tie.sim_id = eso.sim_id
			AND eso_tie.rank = eso.rank
			AND eso_tie.deleted_at IS NULL
	) AS is_tied
FROM derived.entry_simulation_outcomes eso
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
		MAX(payout_cents)::double precision AS max_payout_cents,
		COUNT(*)::int AS n_entries
	FROM derived.entry_simulation_outcomes
	WHERE run_id = sqlc.arg(run_id)::text
		AND deleted_at IS NULL
	GROUP BY sim_id
),
entry_sims AS (
	SELECT
		eso.sim_id,
		eso.payout_cents::double precision AS payout_cents,
		eso.points_scored::double precision AS points_scored,
		CASE
			WHEN ps.max_payout_cents > 0 THEN (eso.payout_cents::double precision / ps.max_payout_cents)
			ELSE 0.0::double precision
		END AS normalized_payout
	FROM derived.entry_simulation_outcomes eso
	JOIN per_sim ps ON ps.sim_id = eso.sim_id
	WHERE eso.run_id = sqlc.arg(run_id)::text
		AND eso.entry_name = sqlc.arg(entry_name)::text
		AND eso.deleted_at IS NULL
)
SELECT
	COUNT(*)::int AS total_simulations,
	COALESCE(AVG(payout_cents), 0.0)::double precision AS mean_payout_cents,
	COALESCE(AVG(points_scored), 0.0)::double precision AS mean_points,
	COALESCE(AVG(normalized_payout), 0.0)::double precision AS mean_normalized_payout,
	COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY payout_cents), 0.0)::double precision AS p50_payout_cents,
	COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY payout_cents), 0.0)::double precision AS p90_payout_cents
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
		t.tournament_id AS core_tournament_id
	FROM core.teams t
	WHERE t.id = sqlc.arg(team_id)::uuid
		AND t.deleted_at IS NULL
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
	FROM derived.simulated_teams st
	WHERE st.team_id = sqlc.arg(team_id)::uuid
		AND st.tournament_id = (SELECT core_tournament_id FROM team_ctx)
		AND st.deleted_at IS NULL
	GROUP BY st.team_id, round_name
)
SELECT
	t.id as team_id,
	s.name as school_name,
	t.seed,
	t.region,
	ks.net_rtg as kenpom_net,
	COUNT(DISTINCT st.sim_id)::int as total_sims,
	AVG(st.wins)::float as avg_wins,
	AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float as avg_points,
	jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN core.team_kenpom_stats ks ON ks.team_id = t.id AND ks.deleted_at IS NULL
JOIN valid v ON true


JOIN derived.simulated_teams st
	ON st.team_id = t.id
	AND st.tournament_id = t.tournament_id
	AND st.deleted_at IS NULL
	LEFT JOIN round_distribution rd ON rd.team_id = t.id
WHERE t.id = sqlc.arg(team_id)::uuid
	AND t.deleted_at IS NULL
GROUP BY t.id, s.name, t.seed, t.region, ks.net_rtg;
