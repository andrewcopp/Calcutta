-- name: CreateUser :exec
INSERT INTO core.users (id, email, first_name, last_name, status, password_hash, external_provider, external_provider_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, status, password_hash, external_provider, external_provider_id, created_at, updated_at, deleted_at
FROM core.users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, status, password_hash, external_provider, external_provider_id, created_at, updated_at, deleted_at
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
  external_provider = $7,
  external_provider_id = $8,
  updated_at = $9,
  deleted_at = $10
WHERE id = $1;

-- name: GetUserByExternalProvider :one
SELECT id, email, first_name, last_name, status, password_hash, external_provider, external_provider_id, created_at, updated_at, deleted_at
FROM core.users
WHERE external_provider = $1 AND external_provider_id = $2 AND deleted_at IS NULL;

-- name: GetUsersByIDs :many
SELECT id, email, first_name, last_name, status, password_hash, external_provider, external_provider_id, created_at, updated_at, deleted_at
FROM core.users
WHERE id = ANY(@ids::text[]) AND deleted_at IS NULL;
