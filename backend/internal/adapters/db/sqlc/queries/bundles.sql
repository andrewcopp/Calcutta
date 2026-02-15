-- Bundles / bundle uploads

-- name: UpsertBundleUpload :one
INSERT INTO core.bundle_uploads (filename, sha256, size_bytes, archive, status)
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

-- name: GetBundleUploadStatus :one
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
FROM core.bundle_uploads
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: ClaimNextBundleUpload :one
WITH candidate AS (
	SELECT id
	FROM core.bundle_uploads
	WHERE deleted_at IS NULL
	  AND core.bundle_uploads.finished_at IS NULL
	  AND (
		core.bundle_uploads.status = 'pending'
		OR (core.bundle_uploads.status = 'running' AND core.bundle_uploads.started_at IS NOT NULL AND core.bundle_uploads.started_at < sqlc.arg(stale_before))
	  )
	ORDER BY created_at ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE core.bundle_uploads bu
SET status = 'running',
	started_at = sqlc.arg(now),
	finished_at = NULL,
	error_message = NULL,
	import_report = NULL,
	verify_report = NULL,
	updated_at = NOW()
FROM candidate c
WHERE bu.id = c.id
RETURNING bu.id;

-- name: GetBundleUploadArchive :one
SELECT archive
FROM core.bundle_uploads
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: MarkBundleUploadSucceeded :exec
UPDATE core.bundle_uploads
SET status = 'succeeded',
	finished_at = NOW(),
	import_report = sqlc.arg(import_report),
	verify_report = sqlc.arg(verify_report),
	updated_at = NOW()
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;

-- name: MarkBundleUploadFailed :exec
UPDATE core.bundle_uploads
SET status = 'failed',
	finished_at = NOW(),
	error_message = sqlc.arg(error_message),
	updated_at = NOW()
WHERE id = sqlc.arg(upload_id)::uuid AND deleted_at IS NULL;
