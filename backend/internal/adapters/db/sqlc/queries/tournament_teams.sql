-- name: ListTeamsByTournamentID :many
SELECT
  tt.id,
  tt.tournament_id,
  tt.school_id,
  tt.seed,
  tt.region,
  tt.byes,
  tt.wins,
  tt.eliminated,
  tt.created_at,
  tt.updated_at,
  kps.net_rtg,
  kps.o_rtg,
  kps.d_rtg,
  kps.adj_t,
  s.name AS school_name
FROM core.teams tt
LEFT JOIN core.team_kenpom_stats kps ON kps.team_id = tt.id AND kps.deleted_at IS NULL
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL
ORDER BY tt.seed ASC;

-- name: GetTeamByID :one
SELECT
  tt.id,
  tt.tournament_id,
  tt.school_id,
  tt.seed,
  tt.region,
  tt.byes,
  tt.wins,
  tt.eliminated,
  tt.created_at,
  tt.updated_at,
  kps.net_rtg,
  kps.o_rtg,
  kps.d_rtg,
  kps.adj_t,
  s.name AS school_name
FROM core.teams tt
LEFT JOIN core.team_kenpom_stats kps ON kps.team_id = tt.id AND kps.deleted_at IS NULL
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE tt.id = $1 AND tt.deleted_at IS NULL;

-- name: UpdateTeam :exec
UPDATE core.teams
SET wins = $1,
    byes = $2,
    eliminated = $3,
    updated_at = NOW()
WHERE id = $4 AND deleted_at IS NULL;

-- name: CreateTeam :exec
INSERT INTO core.teams (
  id,
  tournament_id,
  school_id,
  seed,
  region,
  byes,
  wins,
  eliminated,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetWinningTeam :one
SELECT
  tt.id,
  tt.tournament_id,
  tt.school_id,
  tt.seed,
  tt.region,
  tt.byes,
  tt.wins,
  tt.eliminated,
  tt.created_at,
  tt.updated_at,
  kps.net_rtg,
  kps.o_rtg,
  kps.d_rtg,
  kps.adj_t
FROM core.teams tt
LEFT JOIN core.team_kenpom_stats kps ON kps.team_id = tt.id AND kps.deleted_at IS NULL
WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL
ORDER BY tt.wins DESC
LIMIT 1;
