-- name: ListCalcuttas :many
SELECT id, tournament_id, owner_id, name, created_at, updated_at
FROM calcuttas
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCalcuttaByID :one
SELECT id, tournament_id, owner_id, name, created_at, updated_at
FROM calcuttas
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateCalcutta :exec
INSERT INTO calcuttas (id, tournament_id, owner_id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateCalcutta :execrows
UPDATE calcuttas
SET tournament_id = $1,
    owner_id = $2,
    name = $3,
    updated_at = $4
WHERE id = $5 AND deleted_at IS NULL;

-- name: DeleteCalcutta :execrows
UPDATE calcuttas
SET deleted_at = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;

-- name: GetCalcuttasByTournament :many
SELECT id, tournament_id, owner_id, name, created_at, updated_at, deleted_at
FROM calcuttas
WHERE tournament_id = $1 AND deleted_at IS NULL;
