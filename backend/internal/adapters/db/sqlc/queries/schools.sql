-- name: ListSchools :many
SELECT id, name, created_at, updated_at
FROM core.schools
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: GetSchoolByID :one
SELECT id, name, created_at, updated_at
FROM core.schools
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSchool :exec
INSERT INTO core.schools (id, name, slug, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);
