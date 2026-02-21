-- name: ListCalcuttas :many
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_bid_points, budget_points, visibility, created_at, updated_at
FROM core.calcuttas
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCalcuttaByID :one
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_bid_points, budget_points, visibility, created_at, updated_at
FROM core.calcuttas
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateCalcutta :exec
INSERT INTO core.calcuttas (id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_bid_points, budget_points, visibility, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: UpdateCalcutta :execrows
UPDATE core.calcuttas
SET tournament_id = $1,
    owner_id = $2,
    name = $3,
    min_teams = $4,
    max_teams = $5,
    max_bid_points = $6,
    budget_points = $7,
    visibility = $8,
    updated_at = $9
WHERE id = $10 AND deleted_at IS NULL;

-- name: GetCalcuttasByTournament :many
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_bid_points, budget_points, visibility, created_at, updated_at, deleted_at
FROM core.calcuttas
WHERE tournament_id = $1 AND deleted_at IS NULL;

-- name: ListCalcuttasByUserID :many
SELECT DISTINCT c.id, c.tournament_id, c.owner_id, c.created_by, c.name, c.min_teams, c.max_teams, c.max_bid_points, c.budget_points, c.visibility, c.created_at, c.updated_at
FROM core.calcuttas c
WHERE c.deleted_at IS NULL
  AND (c.owner_id = $1
       OR EXISTS (SELECT 1 FROM core.entries e
                  WHERE e.calcutta_id = c.id AND e.user_id = $1 AND e.deleted_at IS NULL)
       OR EXISTS (SELECT 1 FROM core.calcutta_invitations ci
                  WHERE ci.calcutta_id = c.id AND ci.user_id = $1 AND ci.status = 'pending' AND ci.deleted_at IS NULL))
ORDER BY c.created_at DESC;
