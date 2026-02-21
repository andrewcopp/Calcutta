BEGIN;

-- 6. Drop composite index for HasPermission hot path
DROP INDEX IF EXISTS core.idx_grants_user_scope_active;

-- 5. Drop unique constraints on active grants
DROP INDEX IF EXISTS core.uq_grants_user_permission_scope_active;
DROP INDEX IF EXISTS core.uq_grants_user_label_scope_active;

-- 4. Remove user_manager label and its permissions
DELETE FROM core.label_permissions
WHERE label_id = 'b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e';

DELETE FROM core.labels
WHERE id = 'b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e';

-- 3. Remove granted_by from grants
DROP INDEX IF EXISTS core.idx_grants_granted_by;

ALTER TABLE core.grants
    DROP CONSTRAINT IF EXISTS grants_granted_by_fkey;

ALTER TABLE core.grants
    DROP COLUMN IF EXISTS granted_by;

-- 2. Remove soft-delete prevention on calcuttas
DROP TRIGGER IF EXISTS trg_calcuttas_prevent_soft_delete ON core.calcuttas;
DROP FUNCTION IF EXISTS core.prevent_calcutta_soft_delete();

-- 1. Remove created_by from calcuttas
DROP TRIGGER IF EXISTS trg_calcuttas_immutable_created_by ON core.calcuttas;
DROP FUNCTION IF EXISTS core.immutable_created_by();

DROP INDEX IF EXISTS core.idx_calcuttas_created_by;

ALTER TABLE core.calcuttas
    DROP CONSTRAINT IF EXISTS calcuttas_created_by_fkey;

ALTER TABLE core.calcuttas
    DROP COLUMN IF EXISTS created_by;

COMMIT;
