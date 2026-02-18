-- Revert label renames:
-- site_admin → global_admin
-- tournament_admin → tournament_operator
-- calcutta_admin → calcutta_owner

UPDATE core.labels
SET key = 'global_admin',
    description = 'All permissions (admin)'
WHERE key = 'site_admin'
  AND deleted_at IS NULL;

UPDATE core.labels
SET key = 'tournament_operator',
    description = 'Tournament operations'
WHERE key = 'tournament_admin'
  AND deleted_at IS NULL;

UPDATE core.labels
SET key = 'calcutta_owner',
    description = 'Manage a specific calcutta'
WHERE key = 'calcutta_admin'
  AND deleted_at IS NULL;

-- Remove entry.write permission linkages
DELETE FROM core.label_permissions
WHERE permission_id = (SELECT id FROM core.permissions WHERE key = 'entry.write' AND deleted_at IS NULL);

-- Soft delete entry.write permission
UPDATE core.permissions
SET deleted_at = NOW()
WHERE key = 'entry.write'
  AND deleted_at IS NULL;
