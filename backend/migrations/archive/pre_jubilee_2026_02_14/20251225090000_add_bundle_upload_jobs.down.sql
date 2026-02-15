DROP INDEX IF EXISTS idx_bundle_uploads_status_created_at;

ALTER TABLE bundle_uploads DROP CONSTRAINT IF EXISTS bundle_uploads_status_check;

ALTER TABLE bundle_uploads
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS finished_at,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS status;
