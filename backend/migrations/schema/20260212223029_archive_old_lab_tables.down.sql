-- Rollback: Drop archive schema (FK constraints cannot be easily restored)

DROP SCHEMA IF EXISTS archive;

-- Note: FK constraints dropped in the up migration are NOT restored.
-- This is intentional - the constraints referenced tables that will be archived.
-- If you need to restore them, you would need to manually recreate them.
