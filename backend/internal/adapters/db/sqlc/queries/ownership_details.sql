-- name: ListOwnershipDetailsByPortfolioID :many
SELECT
    od.id::text AS id,
    od.portfolio_id,
    od.team_id,
    od.ownership_percentage::float8 AS ownership_percentage,
    od.actual_returns::float8 AS actual_returns,
    od.expected_returns::float8 AS expected_returns,
    od.created_at::timestamptz AS created_at,
    od.updated_at::timestamptz AS updated_at,
    od.deleted_at::timestamptz AS deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.is_eliminated,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    s.name AS school_name
FROM derived.ownership_details od
JOIN core.teams tt ON od.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE od.portfolio_id = $1 AND od.deleted_at IS NULL;

-- name: ListOwnershipDetailsByPortfolioIDs :many
SELECT
    od.id::text AS id,
    od.portfolio_id,
    od.team_id,
    od.ownership_percentage::float8 AS ownership_percentage,
    od.actual_returns::float8 AS actual_returns,
    od.expected_returns::float8 AS expected_returns,
    od.created_at::timestamptz AS created_at,
    od.updated_at::timestamptz AS updated_at,
    od.deleted_at::timestamptz AS deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.is_eliminated,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    s.name AS school_name
FROM derived.ownership_details od
JOIN core.teams tt ON od.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE od.portfolio_id = ANY(@portfolio_ids::uuid[]) AND od.deleted_at IS NULL;
