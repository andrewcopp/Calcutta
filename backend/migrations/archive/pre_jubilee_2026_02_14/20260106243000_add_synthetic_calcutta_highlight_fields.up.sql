ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    ADD COLUMN IF NOT EXISTS highlighted_snapshot_entry_id UUID;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    ADD COLUMN IF NOT EXISTS notes TEXT;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    ADD COLUMN IF NOT EXISTS metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_derived_synthetic_calcuttas_highlighted_snapshot_entry_id'
    ) THEN
        ALTER TABLE derived.synthetic_calcuttas
            ADD CONSTRAINT fk_derived_synthetic_calcuttas_highlighted_snapshot_entry_id
            FOREIGN KEY (highlighted_snapshot_entry_id)
            REFERENCES core.calcutta_snapshot_entries(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_highlighted_snapshot_entry_id
ON derived.synthetic_calcuttas(highlighted_snapshot_entry_id)
WHERE deleted_at IS NULL;
