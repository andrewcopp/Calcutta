-- name: ListSchools :many
SELECT id, name, created_at, updated_at
FROM core.schools
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: GetSchoolByID :one
SELECT id, name, created_at, updated_at
FROM core.schools
WHERE id = $1 AND deleted_at IS NULL;
