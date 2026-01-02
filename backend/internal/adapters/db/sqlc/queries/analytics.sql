-- name: GetSeedAnalytics :many
WITH entry_bids AS (
  SELECT
    cet.team_id,
    ce.calcutta_id,
    cet.bid_points::float AS bid_points,
    SUM(cet.bid_points::float) OVER (
      PARTITION BY ce.calcutta_id, cet.team_id
    ) AS team_total_bid_points,
    tt.wins,
    tt.byes
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_bid_points > 0 THEN
            core.calcutta_points_for_progress(calcutta_id, wins, byes)::float
            * (bid_points / team_total_bid_points)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(bid_points), 0)::float AS total_investment
  FROM entry_bids
  GROUP BY team_id
)
SELECT
  tt.seed,
  COALESCE(SUM(ta.total_points), 0)::float AS total_points,
  COALESCE(SUM(ta.total_investment), 0)::float AS total_investment,
  COUNT(DISTINCT tt.id)::int AS team_count
FROM core.teams tt
LEFT JOIN team_agg ta ON ta.team_id = tt.id
WHERE tt.deleted_at IS NULL
GROUP BY tt.seed
ORDER BY tt.seed;

-- name: GetRegionAnalytics :many
WITH entry_bids AS (
  SELECT
    cet.team_id,
    ce.calcutta_id,
    cet.bid_points::float AS bid_points,
    SUM(cet.bid_points::float) OVER (
      PARTITION BY ce.calcutta_id, cet.team_id
    ) AS team_total_bid_points,
    tt.wins,
    tt.byes
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_bid_points > 0 THEN
            core.calcutta_points_for_progress(calcutta_id, wins, byes)::float
            * (bid_points / team_total_bid_points)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(bid_points), 0)::float AS total_investment
  FROM entry_bids
  GROUP BY team_id
)
SELECT
  tt.region,
  COALESCE(SUM(ta.total_points), 0)::float AS total_points,
  COALESCE(SUM(ta.total_investment), 0)::float AS total_investment,
  COUNT(DISTINCT tt.id)::int AS team_count
FROM core.teams tt
LEFT JOIN team_agg ta ON ta.team_id = tt.id
WHERE tt.deleted_at IS NULL
GROUP BY tt.region
ORDER BY tt.region;

-- name: GetTeamAnalytics :many
WITH entry_bids AS (
  SELECT
    cet.team_id,
    ce.calcutta_id,
    cet.bid_points::float AS bid_points,
    SUM(cet.bid_points::float) OVER (
      PARTITION BY ce.calcutta_id, cet.team_id
    ) AS team_total_bid_points,
    tt.school_id,
    tt.wins,
    tt.byes
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    school_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_bid_points > 0 THEN
            core.calcutta_points_for_progress(calcutta_id, wins, byes)::float
            * (bid_points / team_total_bid_points)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(bid_points), 0)::float AS total_investment
  FROM entry_bids
  GROUP BY team_id, school_id
)
SELECT
  s.id AS school_id,
  s.name AS school_name,
  COALESCE(SUM(ta.total_points), 0)::float AS total_points,
  COALESCE(SUM(ta.total_investment), 0)::float AS total_investment,
  COUNT(DISTINCT tt.id)::int AS appearances,
  COALESCE(SUM(tt.seed), 0)::int AS total_seed
FROM core.schools s
LEFT JOIN core.teams tt ON tt.school_id = s.id AND tt.deleted_at IS NULL
LEFT JOIN team_agg ta ON ta.team_id = tt.id
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.name
HAVING COUNT(DISTINCT tt.id) > 0
ORDER BY total_points DESC;

-- name: GetBestCareers :many
WITH latest_calcutta AS (
  SELECT c.id AS calcutta_id
  FROM core.calcuttas c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  WHERE c.deleted_at IS NULL
  ORDER BY
    seas.year DESC,
    c.created_at DESC
  LIMIT 1
),
latest_entries AS (
  SELECT DISTINCT TRIM(ce.name) AS entry_name
  FROM core.entries ce
  JOIN latest_calcutta lc ON lc.calcutta_id = ce.calcutta_id
  WHERE ce.deleted_at IS NULL
    AND TRIM(ce.name) <> ''
),
entry_points AS (
  SELECT
    c.id AS calcutta_id,
    ce.id AS entry_id,
    ce.created_at AS entry_created_at,
    TRIM(ce.name) AS entry_name,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_bid_points > 0 THEN
            core.calcutta_points_for_progress(ce.calcutta_id, tt.wins, tt.byes)::float
            * (cet.bid_points::float / team_total_bid_points)
          ELSE 0
        END
      ),
      0
    )::float AS total_points
  FROM core.entries ce
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  LEFT JOIN core.entry_teams cet ON cet.entry_id = ce.id AND cet.deleted_at IS NULL
  LEFT JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  LEFT JOIN LATERAL (
    SELECT
      SUM(cet2.bid_points::float) AS team_total_bid_points
    FROM core.entry_teams cet2
    JOIN core.entries ce2 ON ce2.id = cet2.entry_id AND ce2.deleted_at IS NULL
    WHERE ce2.calcutta_id = ce.calcutta_id
      AND cet2.team_id = cet.team_id
      AND cet2.deleted_at IS NULL
  ) team_bids ON true
  WHERE ce.deleted_at IS NULL
  GROUP BY c.id, ce.id, ce.created_at, ce.name
),
group_stats AS (
  SELECT
    calcutta_id,
    total_points,
    COUNT(*)::int AS tie_size,
    DENSE_RANK() OVER (PARTITION BY calcutta_id ORDER BY total_points DESC) AS group_rank
  FROM entry_points
  GROUP BY calcutta_id, total_points
),
position_groups AS (
  SELECT
    gs.*, 
    1 + COALESCE(
      SUM(tie_size) OVER (PARTITION BY calcutta_id ORDER BY group_rank ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING),
      0
    ) AS start_position
  FROM group_stats gs
),
enriched AS (
  SELECT
    ep.calcutta_id,
    ep.entry_id,
    ep.entry_name,
    ep.total_points,
    pg.tie_size,
    pg.start_position AS finish_position,
    ROW_NUMBER() OVER (
      PARTITION BY ep.calcutta_id, ep.total_points
      ORDER BY ep.entry_created_at DESC, ep.entry_id ASC
    )::int AS tie_index
  FROM entry_points ep
  JOIN position_groups pg
    ON pg.calcutta_id = ep.calcutta_id AND pg.total_points = ep.total_points
  WHERE ep.entry_name <> ''
),
payout_calc AS (
  SELECT
    e.*,
    COALESCE(gp.group_payout_cents, 0)::int AS group_payout_cents,
    (COALESCE(gp.group_payout_cents, 0)::int / NULLIF(e.tie_size, 0))::int AS base_payout_cents,
    (COALESCE(gp.group_payout_cents, 0)::int % NULLIF(e.tie_size, 0))::int AS remainder_cents
  FROM enriched e
  LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(cp.amount_cents), 0)::int AS group_payout_cents
    FROM core.payouts cp
    WHERE cp.calcutta_id = e.calcutta_id
      AND cp.deleted_at IS NULL
      AND cp.position >= e.finish_position
      AND cp.position < (e.finish_position + e.tie_size)
  ) gp ON true
),
entry_results AS (
  SELECT
    pc.entry_name,
    pc.calcutta_id,
    pc.finish_position,
    (
      pc.base_payout_cents + CASE WHEN pc.tie_index <= pc.remainder_cents THEN 1 ELSE 0 END
    )::int AS payout_cents
  FROM payout_calc pc
),
per_calcutta AS (
  SELECT
    entry_name,
    calcutta_id,
    MIN(finish_position)::int AS finish_position,
    SUM(payout_cents)::int AS payout_cents
  FROM entry_results
  GROUP BY entry_name, calcutta_id
),
paid_positions AS (
  SELECT
    cp.calcutta_id,
    MAX(cp.position)::int AS max_paid_position
  FROM core.payouts cp
  WHERE cp.deleted_at IS NULL
    AND cp.amount_cents > 0
  GROUP BY cp.calcutta_id
),
career_agg AS (
  SELECT
    entry_name,
    COUNT(*)::int AS years,
    MIN(finish_position)::int AS best_finish,
    SUM(CASE WHEN finish_position = 1 THEN 1 ELSE 0 END)::int AS wins,
    SUM(CASE WHEN finish_position <= 3 THEN 1 ELSE 0 END)::int AS podiums,
    SUM(CASE WHEN finish_position <= COALESCE(pp.max_paid_position, 0) THEN 1 ELSE 0 END)::int AS in_the_moneys,
    SUM(CASE WHEN finish_position <= 10 THEN 1 ELSE 0 END)::int AS top_10s,
    SUM(payout_cents)::int AS career_earnings_cents
  FROM per_calcutta pc
  LEFT JOIN paid_positions pp ON pp.calcutta_id = pc.calcutta_id
  GROUP BY entry_name
)
SELECT
  entry_name,
  years,
  best_finish,
  wins,
  podiums,
  in_the_moneys,
  top_10s,
  career_earnings_cents,
  EXISTS (
    SELECT 1
    FROM latest_entries le
    WHERE le.entry_name = career_agg.entry_name
  ) AS active_in_latest_calcutta
FROM career_agg
ORDER BY
  (career_earnings_cents::float / NULLIF(years, 0)) DESC,
  wins DESC,
  podiums DESC,
  in_the_moneys DESC,
  top_10s DESC,
  entry_name ASC
LIMIT $1::int;

-- name: GetBestInvestmentBids :many
WITH bid_points AS (
  SELECT
    t.name AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS calcutta_id,
    ce.id AS entry_id,
    ce.name AS entry_name,
    tt.id AS team_id,
    s.name AS school_name,
    tt.seed,
    core.calcutta_points_for_progress(c.id, tt.wins, tt.byes)::float AS team_points,
    cet.bid_points::float AS bid,
    SUM(cet.bid_points::float) OVER (PARTITION BY c.id, tt.id) AS team_total_bid,
    SUM(cet.bid_points::float) OVER (PARTITION BY c.id) AS calcutta_total_bid
  FROM core.teams tt
  JOIN core.entry_teams cet ON cet.team_id = tt.id AND cet.deleted_at IS NULL
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
),
calcutta_points AS (
  SELECT
    distinct_teams.calcutta_id,
    SUM(distinct_teams.team_points) AS calcutta_total_points
  FROM (
    SELECT DISTINCT calcutta_id, team_id, team_points
    FROM bid_points
  ) distinct_teams
  GROUP BY distinct_teams.calcutta_id
),
calcutta_participants AS (
  SELECT
    bp.calcutta_id,
    COUNT(DISTINCT bp.entry_id)::float AS total_participants
  FROM bid_points bp
  GROUP BY bp.calcutta_id
)
SELECT
  tournament_name,
  tournament_year,
  bp.calcutta_id,
  entry_id,
  entry_name,
  team_id,
  school_name,
  seed,
  bid AS investment,
  CASE WHEN team_total_bid > 0 THEN (bid / team_total_bid) ELSE 0 END::float AS ownership_percentage,
  CASE WHEN team_total_bid > 0 THEN (team_points * (bid / team_total_bid)) ELSE 0 END::float AS raw_returns,
  CASE
    WHEN team_total_bid > 0 AND cp.calcutta_total_points > 0 AND cpar.total_participants > 0 THEN (team_points * (bid / team_total_bid)) * (100.0 * cpar.total_participants) / cp.calcutta_total_points
    ELSE 0
  END::float AS normalized_returns
FROM bid_points bp
JOIN calcutta_points cp ON cp.calcutta_id = bp.calcutta_id
JOIN calcutta_participants cpar ON cpar.calcutta_id = bp.calcutta_id
WHERE bid > 0
  AND team_points > 0
ORDER BY normalized_returns DESC, raw_returns DESC, investment ASC
LIMIT $1::int;

-- name: GetBestEntries :many
WITH entry_points AS (
  SELECT
    t.name AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS calcutta_id,
    ce.id AS entry_id,
    ce.name AS entry_name,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_bid_points > 0 THEN
            core.calcutta_points_for_progress(ce.calcutta_id, tt.wins, tt.byes)::float
            * (cet.bid_points::float / team_total_bid_points)
          ELSE 0
        END
      ),
      0
    )::float AS total_points
  FROM core.entries ce
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  LEFT JOIN core.entry_teams cet ON cet.entry_id = ce.id AND cet.deleted_at IS NULL
  LEFT JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  LEFT JOIN LATERAL (
    SELECT
      SUM(cet2.bid_points::float) AS team_total_bid_points
    FROM core.entry_teams cet2
    JOIN core.entries ce2 ON ce2.id = cet2.entry_id AND ce2.deleted_at IS NULL
    WHERE ce2.calcutta_id = ce.calcutta_id
      AND cet2.team_id = cet.team_id
      AND cet2.deleted_at IS NULL
  ) team_bids ON true
  WHERE ce.deleted_at IS NULL
  GROUP BY t.name, tournament_year, c.id, ce.id, ce.name
),
enriched AS (
  SELECT
    entry_points.*,
    COUNT(*) OVER (PARTITION BY calcutta_id)::int AS total_participants,
    AVG(total_points) OVER (PARTITION BY calcutta_id)::float AS average_returns,
    SUM(total_points) OVER (PARTITION BY calcutta_id)::float AS calcutta_total_returns
  FROM entry_points
)
SELECT
  tournament_name,
  tournament_year,
  calcutta_id,
  entry_id,
  entry_name,
  total_points AS total_returns,
  total_participants,
  average_returns,
  CASE
    WHEN calcutta_total_returns > 0 AND total_participants > 0 THEN total_points * (100.0 * total_participants::float) / calcutta_total_returns
    ELSE 0
  END::float AS normalized_returns
FROM enriched
WHERE total_points > 0
ORDER BY normalized_returns DESC, total_points DESC
LIMIT $1::int;

-- name: GetBestInvestments :many
WITH team_bids AS (
  SELECT
    t.name AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS calcutta_id,
    tt.id AS team_id,
    s.name AS school_name,
    tt.seed,
    tt.region,
    core.calcutta_points_for_progress(c.id, tt.wins, tt.byes)::float AS team_points,
    SUM(cet.bid_points)::float AS total_bid
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
  GROUP BY t.name, tournament_year, c.id, tt.id, s.name, tt.seed, tt.region, team_points
),
calcutta_enriched AS (
  SELECT
    team_bids.*,
    SUM(total_bid) OVER (PARTITION BY calcutta_id)::float AS calcutta_total_bid,
    SUM(team_points) OVER (PARTITION BY calcutta_id)::float AS calcutta_total_points
  FROM team_bids
)
SELECT
  tournament_name,
  tournament_year,
  calcutta_id,
  team_id,
  school_name,
  seed,
  region,
  team_points,
  total_bid,
  calcutta_total_bid,
  calcutta_total_points,
  CASE WHEN calcutta_total_bid > 0 THEN (total_bid / calcutta_total_bid) ELSE 0 END::float AS investment_share,
  CASE WHEN calcutta_total_points > 0 THEN (team_points / calcutta_total_points) ELSE 0 END::float AS points_share,
  CASE WHEN total_bid > 0 THEN (team_points / total_bid) ELSE 0 END::float AS raw_roi,
  CASE
    WHEN total_bid > 0 AND calcutta_total_bid > 0 AND calcutta_total_points > 0 THEN (team_points / total_bid) / (calcutta_total_points / calcutta_total_bid)
    ELSE 0
  END::float AS normalized_roi
FROM calcutta_enriched
WHERE total_bid > 0
  AND team_points > 0
ORDER BY normalized_roi DESC, raw_roi DESC, total_bid ASC
LIMIT $1::int;

-- name: GetSeedInvestmentPoints :many
WITH team_bids AS (
  SELECT
    tt.seed,
    t.name AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS calcutta_id,
    tt.id AS team_id,
    s.name AS school_name,
    SUM(cet.bid_points)::float AS total_bid
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
    AND (tt.byes > 0 OR tt.wins > 0)
  GROUP BY tt.seed, t.name, tournament_year, c.id, tt.id, s.name
)
SELECT
  seed,
  tournament_name,
  tournament_year,
  calcutta_id,
  team_id,
  school_name,
  total_bid,
  SUM(total_bid) OVER (PARTITION BY calcutta_id)::float AS calcutta_total_bid,
  (total_bid / NULLIF(SUM(total_bid) OVER (PARTITION BY calcutta_id), 0))::float AS normalized_bid
FROM team_bids
ORDER BY tournament_year DESC, calcutta_id, seed ASC, school_name ASC;

-- name: GetCalcuttaPredictedInvestment :many
WITH calcutta AS (
  SELECT c.id AS calcutta_id
  FROM core.calcuttas c
  WHERE c.id = $1::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
bronze_calcutta AS (
  SELECT
    bc.id AS bronze_calcutta_id,
    bc.tournament_id AS bronze_tournament_id
  FROM bronze.calcuttas bc
  JOIN calcutta c ON c.calcutta_id = bc.core_calcutta_id
  LIMIT 1
),
bronze_tournament AS (
  SELECT bronze_tournament_id AS id
  FROM bronze_calcutta
),
entry_count AS (
  SELECT COUNT(DISTINCT entry_name) as num_entries
  FROM bronze.entry_bids beb
  WHERE beb.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta)
),
total_pool AS (
  SELECT (COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47)::double precision * 100.0::double precision)::double precision as pool_size
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float AS expected_points
  FROM silver.simulated_tournaments st
  WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
  GROUP BY st.team_id
),
total_expected_points AS (
  SELECT SUM(expected_points) as total_ev
  FROM team_expected_points
)
SELECT
  t.id as team_id,
  t.school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  ((COALESCE(tep.expected_points, 0.0)::double precision / NULLIF((SELECT total_ev FROM total_expected_points)::double precision, 0.0::double precision)) * (SELECT pool_size FROM total_pool))::double precision as rational,
  (COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0)::double precision * (SELECT pool_size FROM total_pool))::double precision as predicted,
  CASE
    WHEN (COALESCE(tep.expected_points, 0.0) / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool) > 0
    THEN ((COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0) * (SELECT pool_size FROM total_pool)) -
      ((COALESCE(tep.expected_points, 0.0) / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool))) /
      ((COALESCE(tep.expected_points, 0.0) / NULLIF((SELECT total_ev FROM total_expected_points), 0)) * (SELECT pool_size FROM total_pool)) * 100
    ELSE 0
  END::double precision as delta
FROM bronze.teams t
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
LEFT JOIN silver.predicted_market_share spms_c
  ON spms_c.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta) AND spms_c.team_id = t.id
LEFT JOIN silver.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT id FROM bronze_tournament)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
ORDER BY predicted DESC, seed ASC;

-- name: GetCalcuttaPredictedReturns :many
WITH calcutta AS (
  SELECT c.id AS calcutta_id
  FROM core.calcuttas c
  WHERE c.id = $1::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
bronze_calcutta AS (
  SELECT bc.tournament_id AS bronze_tournament_id
  FROM bronze.calcuttas bc
  JOIN calcutta c ON c.calcutta_id = bc.core_calcutta_id
  LIMIT 1
),
bronze_tournament AS (
  SELECT bronze_tournament_id AS id
  FROM bronze_calcutta
),
team_win_counts AS (
  SELECT
    st.team_id,
    st.wins,
    COUNT(*) as sim_count
  FROM silver.simulated_tournaments st
  WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
  GROUP BY st.team_id, st.wins
),
team_probabilities AS (
  SELECT
    team_id,
    SUM(sim_count)::float as total_sims,
    SUM(CASE WHEN wins = 0 THEN sim_count ELSE 0 END)::float as win_pi,
    SUM(CASE WHEN wins = 1 THEN sim_count ELSE 0 END)::float as win_r64,
    SUM(CASE WHEN wins = 2 THEN sim_count ELSE 0 END)::float as win_r32,
    SUM(CASE WHEN wins = 3 THEN sim_count ELSE 0 END)::float as win_s16,
    SUM(CASE WHEN wins = 4 THEN sim_count ELSE 0 END)::float as win_e8,
    SUM(CASE WHEN wins = 5 THEN sim_count ELSE 0 END)::float as win_ff,
    SUM(CASE WHEN wins = 6 THEN sim_count ELSE 0 END)::float as win_champ
  FROM team_win_counts
  GROUP BY team_id
),
team_expected_value AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float AS expected_value
  FROM silver.simulated_tournaments st
  WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
  GROUP BY st.team_id
)
SELECT
  t.id as team_id,
  t.school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(tp.win_pi / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_pi,
  COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r64,
  COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_r32,
  COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_s16,
  COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_e8,
  COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_ff,
  COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0.0::double precision), 0.0::double precision)::double precision as prob_champ,
  COALESCE(tev.expected_value, 0.0)::double precision as expected_value
FROM bronze.teams t
LEFT JOIN team_probabilities tp ON t.id = tp.team_id
LEFT JOIN team_expected_value tev ON t.id = tev.team_id
WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
ORDER BY expected_value DESC, seed ASC;

-- name: GetCalcuttaSimulatedEntry :many
WITH calcutta AS (
  SELECT c.id AS calcutta_id
  FROM core.calcuttas c
  WHERE c.id = $1::uuid
    AND c.deleted_at IS NULL
  LIMIT 1
),
bronze_calcutta AS (
  SELECT
    bc.id AS bronze_calcutta_id,
    bc.tournament_id AS bronze_tournament_id,
    bt.season AS season
  FROM bronze.calcuttas bc
  JOIN bronze.tournaments bt ON bt.id = bc.tournament_id
  JOIN calcutta c ON c.calcutta_id = bc.core_calcutta_id
  LIMIT 1
),
bronze_tournament AS (
  SELECT bronze_tournament_id AS id
  FROM bronze_calcutta
),
entry_count AS (
  SELECT COUNT(DISTINCT entry_name) as num_entries
  FROM bronze.entry_bids beb
  WHERE beb.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta)
),
total_pool AS (
  SELECT COALESCE(NULLIF((SELECT num_entries FROM entry_count), 0), 47) * 100.0 as pool_size
),
team_expected_points AS (
  SELECT
    st.team_id,
    AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float AS expected_points
  FROM silver.simulated_tournaments st
  WHERE st.tournament_id = (SELECT id FROM bronze_tournament)
  GROUP BY st.team_id
),
latest_optimization AS (
  SELECT gor.run_id
  FROM gold.optimization_runs gor
  WHERE gor.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta)
  ORDER BY gor.created_at DESC
  LIMIT 1
)
SELECT
  t.id as team_id,
  t.school_name,
  COALESCE(t.seed, 0)::int as seed,
  COALESCE(t.region, '')::text as region,
  COALESCE(tep.expected_points, 0.0)::double precision as expected_points,
  (COALESCE(spms_c.predicted_share, spms_t.predicted_share, 0.0)::double precision * (SELECT pool_size FROM total_pool))::double precision as expected_market,
  COALESCE(reb.recommended_bid_points, 0.0)::double precision as our_bid
FROM bronze.teams t
LEFT JOIN team_expected_points tep ON t.id = tep.team_id
LEFT JOIN silver.predicted_market_share spms_c
  ON spms_c.calcutta_id = (SELECT bronze_calcutta_id FROM bronze_calcutta) AND spms_c.team_id = t.id
LEFT JOIN silver.predicted_market_share spms_t
  ON spms_t.tournament_id = (SELECT id FROM bronze_tournament)
  AND spms_t.calcutta_id IS NULL
  AND spms_t.team_id = t.id
LEFT JOIN latest_optimization lo ON true
LEFT JOIN gold.recommended_entry_bids reb ON reb.run_id = lo.run_id AND reb.team_id = t.id
WHERE t.tournament_id = (SELECT id FROM bronze_tournament)
ORDER BY seed ASC, t.school_name ASC;

-- name: GetSeedVarianceAnalytics :many
WITH team_bids AS (
  SELECT
    tt.seed,
    c.id AS calcutta_id,
    tt.id AS team_id,
    SUM(cet.bid_points)::float AS total_bid,
    core.calcutta_points_for_progress(c.id, tt.wins, tt.byes)::float AS actual_points
  FROM core.entry_teams cet
  JOIN core.entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
  JOIN core.calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
  WHERE cet.deleted_at IS NULL
    AND (tt.byes > 0 OR tt.wins > 0)
  GROUP BY tt.seed, c.id, tt.id
),
normalized AS (
  SELECT
    seed,
    CASE
      WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
      ELSE 0
    END::float AS normalized_bid,
    actual_points
  FROM team_bids
)
SELECT
  seed,
  COALESCE(STDDEV(normalized_bid), 0)::float AS investment_stddev,
  COALESCE(STDDEV(actual_points), 0)::float AS points_stddev,
  COALESCE(AVG(normalized_bid), 0)::float AS investment_mean,
  COALESCE(AVG(actual_points), 0)::float AS points_mean,
  COUNT(*)::int AS team_count
FROM normalized
GROUP BY seed
HAVING COUNT(*) > 1
ORDER BY seed;
