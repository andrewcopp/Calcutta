-- =============================================================================
-- Migration: 20260223000001_rename_bundle_uploads_to_tournament_imports
-- Renames core.bundle_uploads -> core.tournament_imports
-- =============================================================================

-- 1. Rename table
ALTER TABLE core.bundle_uploads RENAME TO tournament_imports;

-- 2. Rename primary key
ALTER TABLE core.tournament_imports RENAME CONSTRAINT bundle_uploads_pkey TO tournament_imports_pkey;

-- 3. Rename check constraint
ALTER TABLE core.tournament_imports RENAME CONSTRAINT ck_core_bundle_uploads_status TO ck_core_tournament_imports_status;

-- 4. Rename indexes
ALTER INDEX uq_bundle_uploads_sha256 RENAME TO uq_tournament_imports_sha256;
ALTER INDEX idx_bundle_uploads_created_at RENAME TO idx_tournament_imports_created_at;
ALTER INDEX idx_bundle_uploads_status_created_at RENAME TO idx_tournament_imports_status_created_at;

-- 5. Rename trigger (DROP + CREATE)
DROP TRIGGER IF EXISTS trg_core_bundle_uploads_updated_at ON core.tournament_imports;
CREATE TRIGGER trg_core_tournament_imports_updated_at
  BEFORE UPDATE ON core.tournament_imports
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
