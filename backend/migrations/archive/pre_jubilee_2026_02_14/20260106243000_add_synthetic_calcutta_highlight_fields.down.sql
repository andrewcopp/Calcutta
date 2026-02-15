DROP INDEX IF EXISTS idx_derived_synthetic_calcuttas_highlighted_snapshot_entry_id;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    DROP CONSTRAINT IF EXISTS fk_derived_synthetic_calcuttas_highlighted_snapshot_entry_id;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    DROP COLUMN IF EXISTS metadata_json;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    DROP COLUMN IF EXISTS notes;

ALTER TABLE IF EXISTS derived.synthetic_calcuttas
    DROP COLUMN IF EXISTS highlighted_snapshot_entry_id;
