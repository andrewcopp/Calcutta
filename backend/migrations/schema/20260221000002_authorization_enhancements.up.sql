BEGIN;

-- =============================================================================
-- 1. Add created_by to calcuttas (permanent creator, distinct from mutable owner_id)
-- =============================================================================

ALTER TABLE core.calcuttas
    ADD COLUMN IF NOT EXISTS created_by UUID;

-- Backfill: set created_by = owner_id for all existing rows
UPDATE core.calcuttas
SET created_by = owner_id
WHERE created_by IS NULL;

-- Now make it NOT NULL
ALTER TABLE core.calcuttas
    ALTER COLUMN created_by SET NOT NULL;

-- FK to users
ALTER TABLE core.calcuttas
    ADD CONSTRAINT calcuttas_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES core.users(id);

-- Index for lookups by creator
CREATE INDEX IF NOT EXISTS idx_calcuttas_created_by
    ON core.calcuttas (created_by);

-- Prevent created_by from ever being changed after initial insert
CREATE OR REPLACE FUNCTION core.immutable_created_by() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.created_by IS DISTINCT FROM NEW.created_by THEN
        RAISE EXCEPTION 'created_by is immutable and cannot be changed';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_calcuttas_immutable_created_by
    BEFORE UPDATE ON core.calcuttas
    FOR EACH ROW EXECUTE FUNCTION core.immutable_created_by();

-- =============================================================================
-- 2. Prevent soft-deletion of calcuttas at the DB level
-- =============================================================================

CREATE OR REPLACE FUNCTION core.prevent_calcutta_soft_delete() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'Calcuttas cannot be deleted';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_calcuttas_prevent_soft_delete
    BEFORE UPDATE ON core.calcuttas
    FOR EACH ROW EXECUTE FUNCTION core.prevent_calcutta_soft_delete();

-- =============================================================================
-- 3. Add granted_by audit column to grants
-- =============================================================================

ALTER TABLE core.grants
    ADD COLUMN IF NOT EXISTS granted_by UUID;

ALTER TABLE core.grants
    ADD CONSTRAINT grants_granted_by_fkey
    FOREIGN KEY (granted_by) REFERENCES core.users(id);

CREATE INDEX IF NOT EXISTS idx_grants_granted_by
    ON core.grants (granted_by);

-- =============================================================================
-- 4. Add user_manager label and wire up permissions
-- =============================================================================

INSERT INTO core.labels (id, key, description)
VALUES ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', 'user_manager', 'Manage user accounts: password resets, invite resends, user admin')
ON CONFLICT (key) DO NOTHING;

-- user_manager gets admin.users.read + admin.users.write
INSERT INTO core.label_permissions (label_id, permission_id)
VALUES
    ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', '4d045f02-704b-475c-bb6a-5a88d742a07c'),  -- admin.users.read
    ('b8e1c2d3-f4a5-4b6c-8d7e-9f0a1b2c3d4e', 'caeb9563-7298-4ef3-b9d2-4c13529baa20')   -- admin.users.write
ON CONFLICT (label_id, permission_id) DO NOTHING;

-- =============================================================================
-- 5. Add unique constraint on active grants to prevent duplicate grants
-- =============================================================================

-- Prevent the same user from getting the same label+scope twice (while active)
CREATE UNIQUE INDEX IF NOT EXISTS uq_grants_user_label_scope_active
    ON core.grants (user_id, label_id, scope_type, scope_id)
    WHERE label_id IS NOT NULL
      AND revoked_at IS NULL
      AND deleted_at IS NULL;

-- Prevent the same user from getting the same permission+scope twice (while active)
CREATE UNIQUE INDEX IF NOT EXISTS uq_grants_user_permission_scope_active
    ON core.grants (user_id, permission_id, scope_type, scope_id)
    WHERE permission_id IS NOT NULL
      AND revoked_at IS NULL
      AND deleted_at IS NULL;

-- =============================================================================
-- 6. Composite index for the HasPermission query hot path
-- =============================================================================

-- The HasPermission query filters on (user_id, scope_type, revoked_at, deleted_at)
-- This partial composite index covers the most common lookup pattern
CREATE INDEX IF NOT EXISTS idx_grants_user_scope_active
    ON core.grants (user_id, scope_type, scope_id)
    WHERE revoked_at IS NULL
      AND deleted_at IS NULL;

COMMIT;
