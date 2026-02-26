-- name: ListInvestmentsByPortfolioID :many
SELECT
    inv.id,
    inv.portfolio_id,
    inv.team_id,
    inv.credits,
    inv.created_at,
    inv.updated_at,
    inv.deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    tt.deleted_at AS team_deleted_at,
    s.name AS school_name
FROM core.investments inv
JOIN core.teams tt ON inv.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE inv.portfolio_id = $1 AND inv.deleted_at IS NULL
ORDER BY inv.created_at DESC;

-- name: ListInvestmentsByPortfolioIDs :many
SELECT
    inv.id,
    inv.portfolio_id,
    inv.team_id,
    inv.credits,
    inv.created_at,
    inv.updated_at,
    inv.deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    tt.deleted_at AS team_deleted_at,
    s.name AS school_name
FROM core.investments inv
JOIN core.teams tt ON inv.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE inv.portfolio_id = ANY(@portfolio_ids::uuid[]) AND inv.deleted_at IS NULL
ORDER BY inv.created_at DESC;

-- name: SoftDeleteInvestmentsByPortfolioID :execrows
UPDATE core.investments
SET deleted_at = $1,
    updated_at = $1
WHERE portfolio_id = $2 AND deleted_at IS NULL;

-- name: CreateInvestment :exec
INSERT INTO core.investments (id, portfolio_id, team_id, credits, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);
