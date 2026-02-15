-- name: CreateUser :exec
INSERT INTO core.users (id, email, first_name, last_name, status, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, status, password_hash, created_at, updated_at, deleted_at
FROM core.users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, status, password_hash, created_at, updated_at, deleted_at
FROM core.users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUser :exec
UPDATE core.users
SET
  email = $2,
  first_name = $3,
  last_name = $4,
  status = $5,
  password_hash = $6,
  updated_at = $7,
  deleted_at = $8
WHERE id = $1;
