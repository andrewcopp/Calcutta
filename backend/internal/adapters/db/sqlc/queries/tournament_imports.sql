-- Tournament imports (formerly bundle_uploads)

-- name: UpsertTournamentImport :one
INSERT INTO core.tournament_imports (filename, sha256, size_bytes, archive, status)
VALUES (sqlc.arg(filename), sqlc.arg(sha256), sqlc.arg(size_bytes), sqlc.arg(archive), 'pending')
ON CONFLICT (sha256) WHERE deleted_at IS NULL
DO UPDATE SET
	filename = EXCLUDED.filename,
	size_bytes = EXCLUDED.size_bytes,
	archive = EXCLUDED.archive,
	status = 'pending',
	started_at = NULL,
	finished_at = NULL,
	error_message = NULL,
	import_report = NULL,
	verify_report = NULL,
	updated_at = NOW()
RETURNING id;

-- name: GetTournamentImportStatus :one
SELECT
	filename,
	sha256,
	size_bytes,
	status,
	started_at,
	finished_at,
	error_message,
	COALESCE(import_report, '{}'::jsonb)::jsonb AS import_report,
	COALESCE(verify_report, '{}'::jsonb)::jsonb AS verify_report
FROM core.tournament_imports
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: ClaimNextTournamentImport :one
WITH candidate AS (
	SELECT id
	FROM core.tournament_imports
	WHERE deleted_at IS NULL
	  AND core.tournament_imports.finished_at IS NULL
	  AND (
		core.tournament_imports.status = 'pending'
		OR (core.tournament_imports.status = 'running' AND core.tournament_imports.started_at IS NOT NULL AND core.tournament_imports.started_at < sqlc.arg(stale_before))
	  )
	ORDER BY created_at ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE core.tournament_imports ti
SET status = 'running',
	started_at = sqlc.arg(now),
	finished_at = NULL,
	error_message = NULL,
	import_report = NULL,
	verify_report = NULL,
	updated_at = NOW()
FROM candidate c
WHERE ti.id = c.id
RETURNING ti.id;

-- name: GetTournamentImportArchive :one
SELECT archive
FROM core.tournament_imports
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: MarkTournamentImportSucceeded :exec
UPDATE core.tournament_imports
SET status = 'succeeded',
	finished_at = NOW(),
	import_report = sqlc.arg(import_report),
	verify_report = sqlc.arg(verify_report),
	updated_at = NOW()
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: MarkTournamentImportFailed :exec
UPDATE core.tournament_imports
SET status = 'failed',
	finished_at = NOW(),
	error_message = sqlc.arg(error_message),
	updated_at = NOW()
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;
