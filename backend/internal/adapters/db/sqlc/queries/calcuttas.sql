-- name: ListCalcuttas :many
SELECT id, tournament_id, owner_id, name, min_teams, max_teams, max_bid, bidding_open, bidding_locked_at, created_at, updated_at
FROM core.calcuttas
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCalcuttaByID :one
SELECT id, tournament_id, owner_id, name, min_teams, max_teams, max_bid, bidding_open, bidding_locked_at, created_at, updated_at
FROM core.calcuttas
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateCalcutta :exec
INSERT INTO core.calcuttas (id, tournament_id, owner_id, name, min_teams, max_teams, max_bid, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateCalcutta :execrows
UPDATE core.calcuttas
SET tournament_id = $1,
    owner_id = $2,
    name = $3,
    min_teams = $4,
    max_teams = $5,
    max_bid = $6,
    bidding_open = $7,
    bidding_locked_at = $8,
    updated_at = $9
WHERE id = $10 AND deleted_at IS NULL;

-- name: DeleteCalcutta :execrows
UPDATE core.calcuttas
SET deleted_at = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;

-- name: GetCalcuttasByTournament :many
SELECT id, tournament_id, owner_id, name, min_teams, max_teams, max_bid, bidding_open, bidding_locked_at, created_at, updated_at, deleted_at
FROM core.calcuttas
WHERE tournament_id = $1 AND deleted_at IS NULL;
