-- name: ListEntryTeamsByEntryID :many
SELECT
    cet.id,
    cet.entry_id,
    cet.team_id,
    cet.bid_points,
    cet.created_at,
    cet.updated_at,
    cet.deleted_at,
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
FROM core.entry_teams cet
JOIN core.teams tt ON cet.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE cet.entry_id = $1 AND cet.deleted_at IS NULL
ORDER BY cet.created_at DESC;

-- name: ListEntryTeamsByEntryIDs :many
SELECT
    cet.id,
    cet.entry_id,
    cet.team_id,
    cet.bid_points,
    cet.created_at,
    cet.updated_at,
    cet.deleted_at,
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
FROM core.entry_teams cet
JOIN core.teams tt ON cet.team_id = tt.id
LEFT JOIN core.schools s ON tt.school_id = s.id
WHERE cet.entry_id = ANY(@entry_ids::uuid[]) AND cet.deleted_at IS NULL
ORDER BY cet.created_at DESC;

-- name: SoftDeleteEntryTeamsByEntryID :execrows
UPDATE core.entry_teams
SET deleted_at = $1,
    updated_at = $1
WHERE entry_id = $2 AND deleted_at IS NULL;

-- name: CreateEntryTeam :exec
INSERT INTO core.entry_teams (id, entry_id, team_id, bid_points, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);
