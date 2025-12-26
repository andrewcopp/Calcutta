-- name: CreateUser :exec
INSERT INTO users (id, email, first_name, last_name, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, password_hash, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, password_hash, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUser :exec
UPDATE users
SET
  email = $2,
  first_name = $3,
  last_name = $4,
  password_hash = $5,
  updated_at = $6,
  deleted_at = $7
WHERE id = $1;
