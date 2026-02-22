-- =============================================================================
-- Migration: 20260223000000_rename_labels_to_roles
-- Renames core.labels -> core.roles, core.label_permissions -> core.role_permissions
-- Updates columns, constraints, indexes, and triggers accordingly
-- =============================================================================

-- 1. Rename tables
ALTER TABLE core.labels RENAME TO roles;
ALTER TABLE core.label_permissions RENAME TO role_permissions;

-- 2. Rename columns
ALTER TABLE core.role_permissions RENAME COLUMN label_id TO role_id;
ALTER TABLE core.grants RENAME COLUMN label_id TO role_id;

-- 3. Rename primary keys
ALTER TABLE core.roles RENAME CONSTRAINT labels_pkey TO roles_pkey;
ALTER TABLE core.role_permissions RENAME CONSTRAINT label_permissions_pkey TO role_permissions_pkey;

-- 4. Rename unique constraints
ALTER TABLE core.roles RENAME CONSTRAINT uq_core_labels_key TO uq_core_roles_key;

-- 5. Rename foreign keys
ALTER TABLE core.role_permissions RENAME CONSTRAINT label_permissions_label_id_fkey TO role_permissions_role_id_fkey;
ALTER TABLE core.role_permissions RENAME CONSTRAINT label_permissions_permission_id_fkey TO role_permissions_permission_id_fkey;
ALTER TABLE core.grants RENAME CONSTRAINT grants_label_id_fkey TO grants_role_id_fkey;

-- 6. Rename unique index on grants (filtered unique index)
ALTER INDEX uq_grants_user_label_scope_active RENAME TO uq_grants_user_role_scope_active;

-- 7. Rename indexes
ALTER INDEX idx_label_permissions_label_id RENAME TO idx_role_permissions_role_id;
ALTER INDEX idx_label_permissions_permission_id RENAME TO idx_role_permissions_permission_id;
ALTER INDEX idx_grants_label_id RENAME TO idx_grants_role_id;

-- 8. Rename triggers (DROP + CREATE, Postgres has no ALTER TRIGGER RENAME)
DROP TRIGGER IF EXISTS trg_core_labels_updated_at ON core.roles;
CREATE TRIGGER trg_core_roles_updated_at
  BEFORE UPDATE ON core.roles
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_label_permissions_updated_at ON core.role_permissions;
CREATE TRIGGER trg_core_role_permissions_updated_at
  BEFORE UPDATE ON core.role_permissions
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
