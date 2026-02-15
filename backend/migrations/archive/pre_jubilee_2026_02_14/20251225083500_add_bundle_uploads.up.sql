CREATE TABLE bundle_uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    archive BYTEA NOT NULL,
    import_report JSONB,
    verify_report JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_bundle_uploads_created_at ON bundle_uploads(created_at);
CREATE UNIQUE INDEX uq_bundle_uploads_sha256 ON bundle_uploads(sha256) WHERE deleted_at IS NULL;
