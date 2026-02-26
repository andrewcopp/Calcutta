-- name: GetBestCareers :many
WITH latest_pool AS (
  SELECT c.id AS pool_id
  FROM core.pools c
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.seasons seas ON seas.id = t.season_id
  WHERE c.deleted_at IS NULL
  ORDER BY
    seas.year DESC,
    c.created_at DESC
  LIMIT 1
),
latest_portfolios AS (
  SELECT DISTINCT TRIM(p.name) AS portfolio_name
  FROM core.portfolios p
  JOIN latest_pool lp ON lp.pool_id = p.pool_id
  WHERE p.deleted_at IS NULL
    AND TRIM(p.name) <> ''
),
portfolio_returns AS (
  SELECT
    c.id AS pool_id,
    p.id AS portfolio_id,
    p.created_at AS portfolio_created_at,
    TRIM(p.name) AS portfolio_name,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_credits > 0 THEN
            core.pool_returns_for_progress(p.pool_id, tt.wins, tt.byes)::float
            * (inv.credits::float / team_total_credits)
          ELSE 0
        END
      ),
      0
    )::float AS total_returns
  FROM core.portfolios p
  JOIN core.pools c ON c.id = p.pool_id AND c.deleted_at IS NULL
  LEFT JOIN core.investments inv ON inv.portfolio_id = p.id AND inv.deleted_at IS NULL
  LEFT JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  LEFT JOIN LATERAL (
    SELECT
      SUM(inv2.credits::float) AS team_total_credits
    FROM core.investments inv2
    JOIN core.portfolios p2 ON p2.id = inv2.portfolio_id AND p2.deleted_at IS NULL
    WHERE p2.pool_id = p.pool_id
      AND inv2.team_id = inv.team_id
      AND inv2.deleted_at IS NULL
  ) team_investments ON true
  WHERE p.deleted_at IS NULL
  GROUP BY c.id, p.id, p.created_at, p.name
),
group_stats AS (
  SELECT
    pool_id,
    total_returns,
    COUNT(*)::int AS tie_size,
    DENSE_RANK() OVER (PARTITION BY pool_id ORDER BY total_returns DESC) AS group_rank
  FROM portfolio_returns
  GROUP BY pool_id, total_returns
),
position_groups AS (
  SELECT
    gs.*,
    1 + COALESCE(
      SUM(tie_size) OVER (PARTITION BY pool_id ORDER BY group_rank ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING),
      0
    ) AS start_position
  FROM group_stats gs
),
enriched AS (
  SELECT
    pr.pool_id,
    pr.portfolio_id,
    pr.portfolio_name,
    pr.total_returns,
    pg.tie_size,
    pg.start_position AS finish_position,
    ROW_NUMBER() OVER (
      PARTITION BY pr.pool_id, pr.total_returns
      ORDER BY pr.portfolio_created_at DESC, pr.portfolio_id ASC
    )::int AS tie_index
  FROM portfolio_returns pr
  JOIN position_groups pg
    ON pg.pool_id = pr.pool_id AND pg.total_returns = pr.total_returns
  WHERE pr.portfolio_name <> ''
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
    WHERE cp.pool_id = e.pool_id
      AND cp.deleted_at IS NULL
      AND cp.position >= e.finish_position
      AND cp.position < (e.finish_position + e.tie_size)
  ) gp ON true
),
portfolio_results AS (
  SELECT
    pc.portfolio_name,
    pc.pool_id,
    pc.finish_position,
    (
      pc.base_payout_cents + CASE WHEN pc.tie_index <= pc.remainder_cents THEN 1 ELSE 0 END
    )::int AS payout_cents
  FROM payout_calc pc
),
per_pool AS (
  SELECT
    portfolio_name,
    pool_id,
    MIN(finish_position)::int AS finish_position,
    SUM(payout_cents)::int AS payout_cents
  FROM portfolio_results
  GROUP BY portfolio_name, pool_id
),
paid_positions AS (
  SELECT
    cp.pool_id,
    MAX(cp.position)::int AS max_paid_position
  FROM core.payouts cp
  WHERE cp.deleted_at IS NULL
    AND cp.amount_cents > 0
  GROUP BY cp.pool_id
),
career_agg AS (
  SELECT
    portfolio_name,
    COUNT(*)::int AS years,
    MIN(finish_position)::int AS best_finish,
    SUM(CASE WHEN finish_position = 1 THEN 1 ELSE 0 END)::int AS wins,
    SUM(CASE WHEN finish_position <= 3 THEN 1 ELSE 0 END)::int AS podiums,
    SUM(CASE WHEN finish_position <= COALESCE(pp.max_paid_position, 0) THEN 1 ELSE 0 END)::int AS in_the_moneys,
    SUM(CASE WHEN finish_position <= 10 THEN 1 ELSE 0 END)::int AS top_10s,
    SUM(payout_cents)::int AS career_earnings_cents
  FROM per_pool pp_data
  LEFT JOIN paid_positions pp ON pp.pool_id = pp_data.pool_id
  GROUP BY portfolio_name
)
SELECT
  portfolio_name,
  years,
  best_finish,
  wins,
  podiums,
  in_the_moneys,
  top_10s,
  career_earnings_cents,
  EXISTS (
    SELECT 1
    FROM latest_portfolios lp
    WHERE lp.portfolio_name = career_agg.portfolio_name
  ) AS active_in_latest_pool
FROM career_agg
ORDER BY
  (career_earnings_cents::float / NULLIF(years, 0)) DESC,
  wins DESC,
  podiums DESC,
  in_the_moneys DESC,
  top_10s DESC,
  portfolio_name ASC
LIMIT $1::int;

-- name: GetBestInvestmentBids :many
WITH investment_credits AS (
  SELECT
    (comp.name || ' (' || seas.year || ')')::text AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS pool_id,
    p.id AS portfolio_id,
    p.name AS portfolio_name,
    tt.id AS team_id,
    s.name AS school_name,
    tt.seed,
    core.pool_returns_for_progress(c.id, tt.wins, tt.byes)::float AS team_points,
    inv.credits::float AS investment,
    SUM(inv.credits::float) OVER (PARTITION BY c.id, tt.id) AS team_total_investment,
    SUM(inv.credits::float) OVER (PARTITION BY c.id) AS pool_total_investment
  FROM core.teams tt
  JOIN core.investments inv ON inv.team_id = tt.id AND inv.deleted_at IS NULL
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.pools c ON c.id = p.pool_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.competitions comp ON comp.id = t.competition_id
  JOIN core.seasons seas ON seas.id = t.season_id
  JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
),
pool_points AS (
  SELECT
    distinct_teams.pool_id,
    SUM(distinct_teams.team_points) AS pool_total_points
  FROM (
    SELECT DISTINCT pool_id, team_id, team_points
    FROM investment_credits
  ) distinct_teams
  GROUP BY distinct_teams.pool_id
),
pool_participants AS (
  SELECT
    ic.pool_id,
    COUNT(DISTINCT ic.portfolio_id)::float AS total_participants
  FROM investment_credits ic
  GROUP BY ic.pool_id
)
SELECT
  tournament_name,
  tournament_year,
  ic.pool_id,
  portfolio_id,
  portfolio_name,
  team_id,
  school_name,
  seed,
  investment,
  CASE WHEN team_total_investment > 0 THEN (investment / team_total_investment) ELSE 0 END::float AS ownership_percentage,
  CASE WHEN team_total_investment > 0 THEN (team_points * (investment / team_total_investment)) ELSE 0 END::float AS raw_returns,
  CASE
    WHEN team_total_investment > 0 AND pp.pool_total_points > 0 AND ppar.total_participants > 0 THEN (team_points * (investment / team_total_investment)) * (100.0 * ppar.total_participants) / pp.pool_total_points
    ELSE 0
  END::float AS normalized_returns
FROM investment_credits ic
JOIN pool_points pp ON pp.pool_id = ic.pool_id
JOIN pool_participants ppar ON ppar.pool_id = ic.pool_id
WHERE investment > 0
  AND team_points > 0
ORDER BY normalized_returns DESC, raw_returns DESC, investment ASC
LIMIT $1::int;

-- name: GetBestEntries :many
WITH portfolio_returns AS (
  SELECT
    (comp.name || ' (' || seas.year || ')')::text AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS pool_id,
    p.id AS portfolio_id,
    p.name AS portfolio_name,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_credits > 0 THEN
            core.pool_returns_for_progress(p.pool_id, tt.wins, tt.byes)::float
            * (inv.credits::float / team_total_credits)
          ELSE 0
        END
      ),
      0
    )::float AS total_returns
  FROM core.portfolios p
  JOIN core.pools c ON c.id = p.pool_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.competitions comp ON comp.id = t.competition_id
  JOIN core.seasons seas ON seas.id = t.season_id
  LEFT JOIN core.investments inv ON inv.portfolio_id = p.id AND inv.deleted_at IS NULL
  LEFT JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  LEFT JOIN LATERAL (
    SELECT
      SUM(inv2.credits::float) AS team_total_credits
    FROM core.investments inv2
    JOIN core.portfolios p2 ON p2.id = inv2.portfolio_id AND p2.deleted_at IS NULL
    WHERE p2.pool_id = p.pool_id
      AND inv2.team_id = inv.team_id
      AND inv2.deleted_at IS NULL
  ) team_investments ON true
  WHERE p.deleted_at IS NULL
  GROUP BY comp.name, tournament_year, c.id, p.id, p.name
),
enriched AS (
  SELECT
    portfolio_returns.*,
    COUNT(*) OVER (PARTITION BY pool_id)::int AS total_participants,
    AVG(total_returns) OVER (PARTITION BY pool_id)::float AS average_returns,
    SUM(total_returns) OVER (PARTITION BY pool_id)::float AS pool_total_returns
  FROM portfolio_returns
)
SELECT
  tournament_name,
  tournament_year,
  pool_id,
  portfolio_id,
  portfolio_name,
  total_returns,
  total_participants,
  average_returns,
  CASE
    WHEN pool_total_returns > 0 AND total_participants > 0 THEN total_returns * (100.0 * total_participants::float) / pool_total_returns
    ELSE 0
  END::float AS normalized_returns
FROM enriched
WHERE total_returns > 0
ORDER BY normalized_returns DESC, total_returns DESC
LIMIT $1::int;

-- name: GetBestInvestments :many
WITH team_investments AS (
  SELECT
    (comp.name || ' (' || seas.year || ')')::text AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS pool_id,
    tt.id AS team_id,
    s.name AS school_name,
    tt.seed,
    tt.region,
    core.pool_returns_for_progress(c.id, tt.wins, tt.byes)::float AS team_points,
    SUM(inv.credits)::float AS total_bid
  FROM core.investments inv
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.pools c ON c.id = p.pool_id AND c.deleted_at IS NULL
  JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
  JOIN core.competitions comp ON comp.id = t.competition_id
  JOIN core.seasons seas ON seas.id = t.season_id
  JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
  GROUP BY comp.name, tournament_year, c.id, tt.id, s.name, tt.seed, tt.region, team_points
),
pool_enriched AS (
  SELECT
    team_investments.*,
    SUM(total_bid) OVER (PARTITION BY pool_id)::float AS pool_total_investment,
    SUM(team_points) OVER (PARTITION BY pool_id)::float AS pool_total_points
  FROM team_investments
)
SELECT
  tournament_name,
  tournament_year,
  pool_id,
  team_id,
  school_name,
  seed,
  region,
  team_points,
  total_bid,
  pool_total_investment,
  pool_total_points,
  CASE WHEN pool_total_investment > 0 THEN (total_bid / pool_total_investment) ELSE 0 END::float AS investment_share,
  CASE WHEN pool_total_points > 0 THEN (team_points / pool_total_points) ELSE 0 END::float AS points_share,
  CASE WHEN total_bid > 0 THEN (team_points / total_bid) ELSE 0 END::float AS raw_roi,
  CASE
    WHEN total_bid > 0 AND pool_total_investment > 0 AND pool_total_points > 0 THEN (team_points / total_bid) / (pool_total_points / pool_total_investment)
    ELSE 0
  END::float AS normalized_roi
FROM pool_enriched
WHERE total_bid > 0
  AND team_points > 0
ORDER BY normalized_roi DESC, raw_roi DESC, total_bid ASC
LIMIT $1::int;
