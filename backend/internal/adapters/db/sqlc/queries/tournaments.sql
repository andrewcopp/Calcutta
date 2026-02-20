-- name: ListTournaments :many
SELECT t.id, (comp.name || ' (' || seas.year || ')')::text AS name, t.rounds,
       t.final_four_top_left, t.final_four_bottom_left, t.final_four_top_right, t.final_four_bottom_right,
       t.starting_at,
       t.created_at, t.updated_at
FROM core.tournaments t
JOIN core.competitions comp ON comp.id = t.competition_id
JOIN core.seasons seas ON seas.id = t.season_id
WHERE t.deleted_at IS NULL
ORDER BY seas.year DESC;

-- name: GetTournamentByID :one
SELECT t.id, (comp.name || ' (' || seas.year || ')')::text AS name, t.rounds,
       t.final_four_top_left, t.final_four_bottom_left, t.final_four_top_right, t.final_four_bottom_right,
       t.starting_at,
       t.created_at, t.updated_at
FROM core.tournaments t
JOIN core.competitions comp ON comp.id = t.competition_id
JOIN core.seasons seas ON seas.id = t.season_id
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: CreateCoreTournament :exec
WITH season AS (
  INSERT INTO core.seasons (year)
  VALUES ($2)
  ON CONFLICT (year) DO UPDATE SET year = EXCLUDED.year
  RETURNING id
),
competition AS (
  INSERT INTO core.competitions (name)
  VALUES ($3)
  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
  RETURNING id
)
INSERT INTO core.tournaments (
  id,
  competition_id,
  season_id,
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
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
FROM competition
CROSS JOIN season;

-- name: ListCompetitions :many
SELECT id, name
FROM core.competitions
WHERE deleted_at IS NULL
ORDER BY name;

-- name: ListSeasons :many
SELECT id, year
FROM core.seasons
WHERE deleted_at IS NULL
ORDER BY year DESC;

-- name: UpdateCoreTournamentStartingAt :execrows
UPDATE core.tournaments
SET starting_at = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;

-- name: UpdateCoreTournamentFinalFour :execrows
UPDATE core.tournaments
SET final_four_top_left = $1,
    final_four_bottom_left = $2,
    final_four_top_right = $3,
    final_four_bottom_right = $4,
    updated_at = $5
WHERE id = $6 AND deleted_at IS NULL;
