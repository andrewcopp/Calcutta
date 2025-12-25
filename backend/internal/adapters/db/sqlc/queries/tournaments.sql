-- name: ListTournaments :many
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
       created_at, updated_at
FROM tournaments
WHERE deleted_at IS NULL
ORDER BY name DESC;

-- name: GetTournamentByID :one
SELECT id, name, rounds,
       final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right,
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
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
