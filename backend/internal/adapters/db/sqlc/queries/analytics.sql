-- name: GetSeedAnalytics :many
WITH portfolio_investments AS (
  SELECT
    inv.team_id,
    p.pool_id,
    inv.credits::float AS credits,
    SUM(inv.credits::float) OVER (
      PARTITION BY p.pool_id, inv.team_id
    ) AS team_total_credits,
    tt.wins,
    tt.byes
  FROM core.investments inv
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_credits > 0 THEN
            core.pool_returns_for_progress(pool_id, wins, byes)::float
            * (credits / team_total_credits)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(credits), 0)::float AS total_investment
  FROM portfolio_investments
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
WITH portfolio_investments AS (
  SELECT
    inv.team_id,
    p.pool_id,
    inv.credits::float AS credits,
    SUM(inv.credits::float) OVER (
      PARTITION BY p.pool_id, inv.team_id
    ) AS team_total_credits,
    tt.wins,
    tt.byes
  FROM core.investments inv
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_credits > 0 THEN
            core.pool_returns_for_progress(pool_id, wins, byes)::float
            * (credits / team_total_credits)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(credits), 0)::float AS total_investment
  FROM portfolio_investments
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
WITH portfolio_investments AS (
  SELECT
    inv.team_id,
    p.pool_id,
    inv.credits::float AS credits,
    SUM(inv.credits::float) OVER (
      PARTITION BY p.pool_id, inv.team_id
    ) AS team_total_credits,
    tt.school_id,
    tt.wins,
    tt.byes
  FROM core.investments inv
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
),
team_agg AS (
  SELECT
    team_id,
    school_id,
    COALESCE(
      SUM(
        CASE
          WHEN team_total_credits > 0 THEN
            core.pool_returns_for_progress(pool_id, wins, byes)::float
            * (credits / team_total_credits)
          ELSE 0
        END
      ),
      0
    )::float AS total_points,
    COALESCE(SUM(credits), 0)::float AS total_investment
  FROM portfolio_investments
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

-- name: GetSeedInvestmentPoints :many
WITH team_investments AS (
  SELECT
    tt.seed,
    (comp.name || ' (' || seas.year || ')')::text AS tournament_name,
    seas.year::int AS tournament_year,
    c.id AS pool_id,
    tt.id AS team_id,
    s.name AS school_name,
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
    AND (tt.byes > 0 OR tt.wins > 0)
  GROUP BY tt.seed, comp.name, tournament_year, c.id, tt.id, s.name
)
SELECT
  seed,
  tournament_name,
  tournament_year,
  pool_id,
  team_id,
  school_name,
  total_bid,
  SUM(total_bid) OVER (PARTITION BY pool_id)::float AS pool_total_bid,
  (total_bid / NULLIF(SUM(total_bid) OVER (PARTITION BY pool_id), 0))::float AS normalized_bid
FROM team_investments
ORDER BY tournament_year DESC, pool_id, seed ASC, school_name ASC;

-- name: GetSeedVarianceAnalytics :many
WITH team_investments AS (
  SELECT
    tt.seed,
    c.id AS pool_id,
    tt.id AS team_id,
    SUM(inv.credits)::float AS total_bid,
    core.pool_returns_for_progress(c.id, tt.wins, tt.byes)::float AS actual_points
  FROM core.investments inv
  JOIN core.portfolios p ON p.id = inv.portfolio_id AND p.deleted_at IS NULL
  JOIN core.pools c ON c.id = p.pool_id AND c.deleted_at IS NULL
  JOIN core.teams tt ON tt.id = inv.team_id AND tt.deleted_at IS NULL
  WHERE inv.deleted_at IS NULL
    AND (tt.byes > 0 OR tt.wins > 0)
  GROUP BY tt.seed, c.id, tt.id
),
normalized AS (
  SELECT
    seed,
    CASE
      WHEN SUM(total_bid) OVER (PARTITION BY pool_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY pool_id))
      ELSE 0
    END::float AS normalized_bid,
    actual_points
  FROM team_investments
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
