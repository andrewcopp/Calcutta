ALTER TABLE core.entries DROP CONSTRAINT IF EXISTS ck_entries_status;
ALTER TABLE core.entries DROP COLUMN IF EXISTS status;
