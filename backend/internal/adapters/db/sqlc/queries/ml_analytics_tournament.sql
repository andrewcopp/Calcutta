-- name: GetTournamentSimStatsByYear :one
WITH tournament AS (
	SELECT t.id AS tournament_id,
		seas.year AS season
	FROM core.tournaments t
	JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
	WHERE seas.year = $1::int
		AND t.deleted_at IS NULL
	ORDER BY t.created_at DESC
	LIMIT 1
)
SELECT
	t.tournament_id,
	t.season,
	COUNT(DISTINCT st.sim_id)::int as n_sims,
	COUNT(DISTINCT st.team_id)::int as n_teams,
	AVG(st.wins + st.byes)::float as avg_progress,
	MAX(st.wins + st.byes)::int as max_progress
FROM tournament t
JOIN derived.simulated_teams st ON st.tournament_id = t.tournament_id
GROUP BY t.tournament_id, t.season;

-- name: GetTournamentSimStatsByCoreTournamentID :one
WITH tournament_info AS (
	SELECT
		t.id as tournament_id,
		seas.year as season
	FROM core.tournaments t
	JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
	WHERE t.id = sqlc.arg(core_tournament_id)::uuid
		AND t.deleted_at IS NULL
	LIMIT 1
),
sim_stats AS (
	SELECT
		COUNT(DISTINCT sim_id)::int as total_simulations,
		COUNT(DISTINCT team_id)::int as total_teams
	FROM derived.simulated_teams st
	JOIN tournament_info ti ON st.tournament_id = ti.tournament_id
),
prediction_stats AS (
 	SELECT COUNT(*)::int as total_predictions
 	FROM derived.predicted_game_outcomes pgo
 	JOIN tournament_info ti ON pgo.tournament_id = ti.tournament_id
 ),
win_stats AS (
	SELECT
		AVG(wins)::double precision as mean_wins,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY wins)::double precision as median_wins,
		MAX(wins)::int as max_wins
	FROM derived.simulated_teams st
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
	b.simulation_state_id,
	b.n_sims,
	b.seed,
	b.probability_source_key,
	b.created_at
FROM derived.simulated_tournaments b
WHERE b.tournament_id = $1::uuid
	AND b.deleted_at IS NULL
ORDER BY b.created_at DESC;
