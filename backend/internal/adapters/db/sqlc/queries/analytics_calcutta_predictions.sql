-- name: GetCalcuttaPredictedInvestment :many
WITH calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    t.id AS core_tournament_id,
    seas.year AS year
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  WHERE c.id = sqlc.arg(calcutta_id)::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
entry_count AS (
  SELECT COUNT(*)::int AS num_entries
  FROM core.entries ce
  JOIN calcutta_ctx cc ON cc.calcutta_id = ce.calcutta_id
  WHERE ce.deleted_at IS NULL
),
total_pool AS (
  SELECT (COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47)::double precision * 100.0::double precision)::double precision AS pool_size
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_points
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  GROUP BY st.team_id
),
total_expected_points AS (
  SELECT SUM(expected_points) AS total_ev
  FROM team_expected_points
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool))::double precision, 0.0::double precision) as rational,
  (spms_t.predicted_share * (SELECT pool_size FROM total_pool))::double precision as predicted,
  CASE
    WHEN ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool)) > 0
    THEN (((spms_t.predicted_share * (SELECT pool_size FROM total_pool)) -
      ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool))) /
      ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool)) * 100)
    ELSE 0
  END::double precision as delta
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
JOIN derived.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
  AND spms_t.deleted_at IS NULL
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY predicted DESC, seed ASC;

-- name: GetCalcuttaPredictedInvestmentByOptimizedEntryID :many
WITH optimized_entry AS (
  SELECT
    oe.simulated_tournament_id
  FROM derived.optimized_entries oe
  WHERE oe.id = sqlc.arg(optimized_entry_id)::uuid
    AND oe.deleted_at IS NULL
  LIMIT 1
),
calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    t.id AS core_tournament_id,
    seas.year AS year
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  WHERE c.id = sqlc.arg(calcutta_id)::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
entry_count AS (
  SELECT COUNT(*)::int AS num_entries
  FROM core.entries ce
  JOIN calcutta_ctx cc ON cc.calcutta_id = ce.calcutta_id
  WHERE ce.deleted_at IS NULL
),
total_pool AS (
  SELECT (COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47)::double precision * 100.0::double precision)::double precision AS pool_size
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_points
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
    AND st.simulated_tournament_id = (SELECT simulated_tournament_id FROM optimized_entry)
  GROUP BY st.team_id
),
total_expected_points AS (
  SELECT SUM(expected_points) AS total_ev
  FROM team_expected_points
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool))::double precision, 0.0::double precision) as rational,
  (spms_t.predicted_share * (SELECT pool_size FROM total_pool))::double precision as predicted,
  CASE
    WHEN ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool)) > 0
    THEN (((spms_t.predicted_share * (SELECT pool_size FROM total_pool)) -
      ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool))) /
      ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool)) * 100)
    ELSE 0
  END::double precision as delta
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
JOIN derived.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
  AND spms_t.deleted_at IS NULL
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY predicted DESC, seed ASC;

-- name: GetCalcuttaPredictedReturns :many
WITH calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    t.id AS core_tournament_id
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  WHERE c.id = sqlc.arg(calcutta_id)::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
team_win_counts AS (
  SELECT
    st.team_id,
    (st.wins + st.byes + 1)::int AS progress,
    COUNT(*) as sim_count
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  GROUP BY st.team_id, progress
),
team_probabilities AS (
  SELECT
    team_id,
    SUM(sim_count)::float as total_sims,
    SUM(CASE WHEN progress >= 2 THEN sim_count ELSE 0 END)::float as win_r64,
    SUM(CASE WHEN progress >= 3 THEN sim_count ELSE 0 END)::float as win_r32,
    SUM(CASE WHEN progress >= 4 THEN sim_count ELSE 0 END)::float as win_s16,
    SUM(CASE WHEN progress >= 5 THEN sim_count ELSE 0 END)::float as win_e8,
    SUM(CASE WHEN progress >= 6 THEN sim_count ELSE 0 END)::float as win_ff,
    SUM(CASE WHEN progress >= 7 THEN sim_count ELSE 0 END)::float as win_champ
  FROM team_win_counts
  GROUP BY team_id
),
team_expected_value AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_value
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  GROUP BY st.team_id
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  0.0::double precision as prob_pi,
  COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r64,
  COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r32,
  COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_s16,
  COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_e8,
  COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_ff,
  COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_champ,
  COALESCE(tev.expected_value, 0.0)::double precision as expected_value
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_probabilities tp ON t.id = tp.team_id
LEFT JOIN team_expected_value tev ON t.id = tev.team_id
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY expected_value DESC, seed ASC;

-- name: GetCalcuttaPredictedReturnsByOptimizedEntryID :many
WITH optimized_entry AS (
  SELECT
    oe.simulated_tournament_id
  FROM derived.optimized_entries oe
  WHERE oe.id = sqlc.arg(optimized_entry_id)::uuid
    AND oe.deleted_at IS NULL
  LIMIT 1
),
calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    t.id AS core_tournament_id
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  WHERE c.id = sqlc.arg(calcutta_id)::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
team_win_counts AS (
  SELECT
    st.team_id,
    (st.wins + st.byes + 1)::int AS progress,
    COUNT(*) as sim_count
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
    AND st.simulated_tournament_id = (SELECT simulated_tournament_id FROM optimized_entry)
  GROUP BY st.team_id, progress
),
team_probabilities AS (
  SELECT
    team_id,
    SUM(sim_count)::float as total_sims,
    SUM(CASE WHEN progress >= 2 THEN sim_count ELSE 0 END)::float as win_r64,
    SUM(CASE WHEN progress >= 3 THEN sim_count ELSE 0 END)::float as win_r32,
    SUM(CASE WHEN progress >= 4 THEN sim_count ELSE 0 END)::float as win_s16,
    SUM(CASE WHEN progress >= 5 THEN sim_count ELSE 0 END)::float as win_e8,
    SUM(CASE WHEN progress >= 6 THEN sim_count ELSE 0 END)::float as win_ff,
    SUM(CASE WHEN progress >= 7 THEN sim_count ELSE 0 END)::float as win_champ
  FROM team_win_counts
  GROUP BY team_id
),
team_expected_value AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_value
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
    AND st.simulated_tournament_id = (SELECT simulated_tournament_id FROM optimized_entry)
  GROUP BY st.team_id
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  0.0::double precision as prob_pi,
  COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r64,
  COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r32,
  COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_s16,
  COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_e8,
  COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_ff,
  COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_champ,
  COALESCE(tev.expected_value, 0.0)::double precision as expected_value
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_probabilities tp ON t.id = tp.team_id
LEFT JOIN team_expected_value tev ON t.id = tev.team_id
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY expected_value DESC, seed ASC;

-- name: GetCalcuttaSimulatedEntry :many
WITH calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    c.budget_points,
    t.id AS core_tournament_id,
    seas.year AS year
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  WHERE c.id = $1::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
entry_count AS (
  SELECT COUNT(*)::int AS num_entries
  FROM core.entries ce
  JOIN calcutta_ctx cc ON cc.calcutta_id = ce.calcutta_id
  WHERE ce.deleted_at IS NULL
),
total_pool AS (
  SELECT COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47)
    * COALESCE((SELECT budget_points FROM calcutta_ctx), 100)::double precision AS pool_size
),
latest_optimized_entry AS (
  SELECT oe.id AS optimized_entry_id,
    oe.simulated_tournament_id,
    oe.bids_json
  FROM derived.optimized_entries oe
  JOIN calcutta_ctx cc ON TRUE
  WHERE oe.deleted_at IS NULL
    AND oe.calcutta_id = cc.calcutta_id
  ORDER BY oe.created_at DESC
  LIMIT 1
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_points
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
    AND st.simulated_tournament_id = (SELECT simulated_tournament_id FROM latest_optimized_entry)
  GROUP BY st.team_id
),
total_expected_points AS (
  SELECT SUM(expected_points) AS total_ev
  FROM team_expected_points
),
our_bids AS (
  SELECT
    (bid->>'team_id')::uuid AS team_id,
    (bid->>'bid_points')::int AS bid_points
  FROM latest_optimized_entry loe,
       jsonb_array_elements(loe.bids_json) AS bid
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(tep.expected_points, 0.0)::double precision as expected_points,
  (spms_t.predicted_share * (SELECT pool_size FROM total_pool))::double precision as expected_market,
  COALESCE(ob.bid_points, 0.0)::double precision as our_bid
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
JOIN derived.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
  AND spms_t.deleted_at IS NULL
LEFT JOIN our_bids ob ON ob.team_id = t.id
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY seed ASC, s.name ASC;

-- name: GetLatestOptimizedEntryIDByCoreCalcuttaID :one
SELECT
	oe.id
FROM derived.optimized_entries oe
WHERE oe.calcutta_id = sqlc.arg(calcutta_id)::uuid
	AND oe.deleted_at IS NULL
ORDER BY oe.created_at DESC
LIMIT 1;

-- name: GetCalcuttaSimulatedEntryByOptimizedEntryID :many
WITH calcutta_ctx AS (
  SELECT
    c.id AS calcutta_id,
    c.budget_points,
    t.id AS core_tournament_id
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  WHERE c.id = sqlc.arg(calcutta_id)::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
entry_count AS (
  SELECT COUNT(*)::int AS num_entries
  FROM core.entries ce
  JOIN calcutta_ctx cc ON cc.calcutta_id = ce.calcutta_id
  WHERE ce.deleted_at IS NULL
),
total_pool AS (
  SELECT COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47)
    * COALESCE((SELECT budget_points FROM calcutta_ctx), 100)::double precision AS pool_size
),
optimized_entry AS (
  SELECT oe.simulated_tournament_id, oe.bids_json
  FROM derived.optimized_entries oe
  WHERE oe.id = sqlc.arg(optimized_entry_id)::uuid
    AND oe.deleted_at IS NULL
  LIMIT 1
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins + 1, st.byes))::float AS expected_points
  FROM derived.simulated_teams st
  WHERE st.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
    AND st.simulated_tournament_id = (SELECT simulated_tournament_id FROM optimized_entry)
  GROUP BY st.team_id
),
total_expected_points AS (
  SELECT SUM(expected_points) AS total_ev
  FROM team_expected_points
),
our_bids AS (
  SELECT
    (bid->>'team_id')::uuid AS team_id,
    (bid->>'bid_points')::int AS bid_points
  FROM optimized_entry oe,
       jsonb_array_elements(oe.bids_json) AS bid
)
SELECT
  t.id as team_id,
  s.name as school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(tep.expected_points, 0.0)::double precision as expected_points,
  (spms_t.predicted_share * (SELECT pool_size FROM total_pool))::double precision as expected_market,
  COALESCE(ob.bid_points, 0.0)::double precision as our_bid
FROM core.teams t
JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
JOIN derived.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
  AND spms_t.deleted_at IS NULL
LEFT JOIN our_bids ob ON ob.team_id = t.id
WHERE t.tournament_id = (SELECT core_tournament_id FROM calcutta_ctx)
  AND t.deleted_at IS NULL
ORDER BY seed ASC, s.name ASC;
