-- name: ListTournaments :many
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
       starting_at,
       created_at, updated_at
FROM core.tournaments
WHERE deleted_at IS NULL
ORDER BY name DESC;

-- name: GetTournamentByID :one
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
       starting_at,
       created_at, updated_at
FROM core.tournaments
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateCoreTournament :exec
WITH season_year AS (
  SELECT COALESCE(
    substring($2 from '([0-9]{4})')::int,
    EXTRACT(YEAR FROM $8)::int,
    EXTRACT(YEAR FROM NOW())::int
  ) AS year
),
season AS (
  INSERT INTO core.seasons (year)
  SELECT year FROM season_year
  ON CONFLICT (year) DO UPDATE SET year = EXCLUDED.year
  RETURNING id
),
competition AS (
  INSERT INTO core.competitions (name)
  VALUES ('NCAA Men''s')
  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
  RETURNING id
)
INSERT INTO core.tournaments (
  id,
  competition_id,
  season_id,
  name,
  import_key,
  rounds,
  final_four_top_left,
  final_four_bottom_left,
  final_four_top_right,
  final_four_bottom_right,
  starting_at,
  created_at,
  updated_at
)
SELECT
  $1,
  competition.id,
  season.id,
  $2,
  public.calcutta_slugify($2) || '-' || left(md5($1), 6),
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10
FROM competition
CROSS JOIN season;

-- name: UpdateCoreTournamentStartingAt :execrows
UPDATE core.tournaments
SET starting_at = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;
