ALTER TABLE core.entries ADD COLUMN status text NOT NULL DEFAULT 'draft';
ALTER TABLE core.entries ADD CONSTRAINT ck_entries_status CHECK (status IN ('draft', 'final'));
