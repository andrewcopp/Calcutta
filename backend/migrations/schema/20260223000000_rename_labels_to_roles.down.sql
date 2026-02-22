-- =============================================================================
-- Rollback: 20260223000000_rename_labels_to_roles
-- =============================================================================

-- 8. Restore triggers
DROP TRIGGER IF EXISTS trg_core_role_permissions_updated_at ON core.role_permissions;
CREATE TRIGGER trg_core_label_permissions_updated_at
  BEFORE UPDATE ON core.role_permissions
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_roles_updated_at ON core.roles;
CREATE TRIGGER trg_core_labels_updated_at
  BEFORE UPDATE ON core.roles
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- 7. Restore indexes
ALTER INDEX idx_role_permissions_role_id RENAME TO idx_label_permissions_label_id;
ALTER INDEX idx_role_permissions_permission_id RENAME TO idx_label_permissions_permission_id;
ALTER INDEX idx_grants_role_id RENAME TO idx_grants_label_id;

-- 6. Restore unique index on grants
ALTER INDEX uq_grants_user_role_scope_active RENAME TO uq_grants_user_label_scope_active;

-- 5. Restore foreign keys
ALTER TABLE core.role_permissions RENAME CONSTRAINT role_permissions_role_id_fkey TO label_permissions_label_id_fkey;
ALTER TABLE core.role_permissions RENAME CONSTRAINT role_permissions_permission_id_fkey TO label_permissions_permission_id_fkey;
ALTER TABLE core.grants RENAME CONSTRAINT grants_role_id_fkey TO grants_label_id_fkey;

-- 4. Restore unique constraints
ALTER TABLE core.roles RENAME CONSTRAINT uq_core_roles_key TO uq_core_labels_key;

-- 3. Restore primary keys
ALTER TABLE core.roles RENAME CONSTRAINT roles_pkey TO labels_pkey;
ALTER TABLE core.role_permissions RENAME CONSTRAINT role_permissions_pkey TO label_permissions_pkey;

-- 2. Restore columns
ALTER TABLE core.grants RENAME COLUMN role_id TO label_id;
ALTER TABLE core.role_permissions RENAME COLUMN role_id TO label_id;

-- 1. Restore tables
ALTER TABLE core.role_permissions RENAME TO label_permissions;
ALTER TABLE core.roles RENAME TO labels;
