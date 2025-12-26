-- name: ListTournaments :many
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
       starting_at,
       created_at, updated_at
FROM tournaments
WHERE deleted_at IS NULL
ORDER BY name DESC;

-- name: GetTournamentByID :one
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
       starting_at,
       created_at, updated_at
FROM tournaments
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateTournament :exec
INSERT INTO tournaments (
  id,
  name,
  rounds,
  final_four_top_left,
  final_four_bottom_left,
  final_four_top_right,
  final_four_bottom_right,
  starting_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateTournamentStartingAt :execrows
UPDATE tournaments
SET starting_at = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;
