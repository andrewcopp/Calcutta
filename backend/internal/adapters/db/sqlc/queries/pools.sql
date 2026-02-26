-- name: ListPools :many
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_investment_credits, budget_credits, visibility, created_at, updated_at
FROM core.pools
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetPoolByID :one
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_investment_credits, budget_credits, visibility, created_at, updated_at
FROM core.pools
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePool :exec
INSERT INTO core.pools (id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_investment_credits, budget_credits, visibility, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: UpdatePool :execrows
UPDATE core.pools
SET tournament_id = $1,
    owner_id = $2,
    name = $3,
    min_teams = $4,
    max_teams = $5,
    max_investment_credits = $6,
    budget_credits = $7,
    visibility = $8,
    updated_at = $9
WHERE id = $10 AND deleted_at IS NULL;

-- name: GetPoolsByTournament :many
SELECT id, tournament_id, owner_id, created_by, name, min_teams, max_teams, max_investment_credits, budget_credits, visibility, created_at, updated_at, deleted_at
FROM core.pools
WHERE tournament_id = $1 AND deleted_at IS NULL;

-- name: ListPoolsByUserID :many
SELECT DISTINCT c.id, c.tournament_id, c.owner_id, c.created_by, c.name, c.min_teams, c.max_teams, c.max_investment_credits, c.budget_credits, c.visibility, c.created_at, c.updated_at
FROM core.pools c
WHERE c.deleted_at IS NULL
  AND (c.owner_id = $1
       OR EXISTS (SELECT 1 FROM core.portfolios e
                  WHERE e.pool_id = c.id AND e.user_id = $1 AND e.deleted_at IS NULL)
       OR EXISTS (SELECT 1 FROM core.pool_invitations ci
                  WHERE ci.pool_id = c.id AND ci.user_id = $1 AND ci.status = 'pending' AND ci.deleted_at IS NULL))
ORDER BY c.created_at DESC;
