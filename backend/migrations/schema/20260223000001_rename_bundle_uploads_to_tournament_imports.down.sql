-- =============================================================================
-- Rollback: 20260223000001_rename_bundle_uploads_to_tournament_imports
-- =============================================================================

-- 5. Restore trigger
DROP TRIGGER IF EXISTS trg_core_tournament_imports_updated_at ON core.tournament_imports;
CREATE TRIGGER trg_core_bundle_uploads_updated_at
  BEFORE UPDATE ON core.tournament_imports
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 4. Restore indexes
ALTER INDEX idx_tournament_imports_status_created_at RENAME TO idx_bundle_uploads_status_created_at;
ALTER INDEX idx_tournament_imports_created_at RENAME TO idx_bundle_uploads_created_at;
ALTER INDEX uq_tournament_imports_sha256 RENAME TO uq_bundle_uploads_sha256;

-- 3. Restore check constraint
ALTER TABLE core.tournament_imports RENAME CONSTRAINT ck_core_tournament_imports_status TO ck_core_bundle_uploads_status;

-- 2. Restore primary key
ALTER TABLE core.tournament_imports RENAME CONSTRAINT tournament_imports_pkey TO bundle_uploads_pkey;

-- 1. Restore table
ALTER TABLE core.tournament_imports RENAME TO bundle_uploads;
