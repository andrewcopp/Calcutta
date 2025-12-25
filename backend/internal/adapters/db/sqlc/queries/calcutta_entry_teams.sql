-- name: ListEntryTeamsByEntryID :many
SELECT
    cet.id,
    cet.entry_id,
    cet.team_id,
    cet.bid,
    cet.created_at,
    cet.updated_at,
    cet.deleted_at,
    tt.id AS tournament_team_id,
    tt.school_id,
    tt.tournament_id,
    tt.seed,
    tt.byes,
    tt.wins,
    tt.created_at AS team_created_at,
    tt.updated_at AS team_updated_at,
    tt.deleted_at AS team_deleted_at,
    s.name AS school_name
FROM calcutta_entry_teams cet
JOIN tournament_teams tt ON cet.team_id = tt.id
LEFT JOIN schools s ON tt.school_id = s.id
WHERE cet.entry_id = $1 AND cet.deleted_at IS NULL
ORDER BY cet.created_at DESC;
