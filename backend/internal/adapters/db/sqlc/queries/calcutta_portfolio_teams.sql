-- name: ListPortfolioTeamsByPortfolioID :many
SELECT
    cpt.id,
    cpt.portfolio_id,
    cpt.team_id,
    cpt.ownership_percentage::float8 AS ownership_percentage,
    cpt.actual_points::float8 AS actual_points,
    cpt.expected_points::float8 AS expected_points,
    cpt.predicted_points::float8 AS predicted_points,
    cpt.created_at::timestamptz AS created_at,
    cpt.updated_at::timestamptz AS updated_at,
    cpt.deleted_at::timestamptz AS deleted_at,
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
FROM core.derived_portfolio_teams cpt
JOIN core.teams tt ON cpt.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE cpt.portfolio_id = $1 AND cpt.deleted_at IS NULL;
