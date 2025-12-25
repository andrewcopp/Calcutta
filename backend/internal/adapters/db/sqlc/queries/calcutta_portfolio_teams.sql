-- name: ListPortfolioTeamsByPortfolioID :many
SELECT
    cpt.id,
    cpt.portfolio_id,
    cpt.team_id,
    cpt.ownership_percentage::float8 AS ownership_percentage,
    cpt.actual_points::float8 AS actual_points,
    cpt.expected_points::float8 AS expected_points,
    cpt.predicted_points::float8 AS predicted_points,
    cpt.created_at,
    cpt.updated_at,
    cpt.deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.eliminated,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    s.name AS school_name
FROM calcutta_portfolio_teams cpt
JOIN tournament_teams tt ON cpt.team_id = tt.id
LEFT JOIN schools s ON tt.school_id = s.id
WHERE cpt.portfolio_id = $1 AND cpt.deleted_at IS NULL;

-- name: UpdatePortfolioTeam :execrows
UPDATE calcutta_portfolio_teams
SET ownership_percentage = $1,
    actual_points = $2,
    expected_points = $3,
    predicted_points = $4,
    updated_at = $5
WHERE id = $6 AND deleted_at IS NULL;

-- name: CreatePortfolioTeam :exec
INSERT INTO calcutta_portfolio_teams (
    id,
    portfolio_id,
    team_id,
    ownership_percentage,
    actual_points,
    expected_points,
    predicted_points,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
