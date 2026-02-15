ALTER TABLE bundle_uploads ADD COLUMN status TEXT;
ALTER TABLE bundle_uploads ADD COLUMN started_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE bundle_uploads ADD COLUMN finished_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE bundle_uploads ADD COLUMN error_message TEXT;

UPDATE bundle_uploads
SET status = CASE
    WHEN import_report IS NOT NULL AND verify_report IS NOT NULL THEN 'succeeded'
    ELSE 'pending'
END
WHERE status IS NULL;

ALTER TABLE bundle_uploads ALTER COLUMN status SET NOT NULL;
ALTER TABLE bundle_uploads ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE bundle_uploads
    ADD CONSTRAINT bundle_uploads_status_check
    CHECK (status IN ('pending', 'running', 'succeeded', 'failed'));

CREATE INDEX idx_bundle_uploads_status_created_at
    ON bundle_uploads(status, created_at)
    WHERE deleted_at IS NULL;
