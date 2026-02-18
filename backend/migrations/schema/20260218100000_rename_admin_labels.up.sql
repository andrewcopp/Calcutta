-- Rename admin labels for clearer hierarchy:
-- global_admin → site_admin
-- tournament_operator → tournament_admin
-- calcutta_owner → calcutta_admin

UPDATE core.labels
SET key = 'site_admin',
    description = 'Full site administration'
WHERE key = 'global_admin'
  AND deleted_at IS NULL;

UPDATE core.labels
SET key = 'tournament_admin',
    description = 'Update tournament game results'
WHERE key = 'tournament_operator'
  AND deleted_at IS NULL;

UPDATE core.labels
SET key = 'calcutta_admin',
    description = 'Manage a specific calcutta'
WHERE key = 'calcutta_owner'
  AND deleted_at IS NULL;

-- Add entry.write permission for calcutta admins
INSERT INTO core.permissions (id, key, description, created_at, updated_at)
VALUES (
  'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
  'entry.write',
  'Edit any entry in a calcutta',
  NOW(),
  NOW()
)
ON CONFLICT DO NOTHING;

-- Link entry.write to calcutta_admin label
INSERT INTO core.label_permissions (label_id, permission_id, created_at, updated_at)
SELECT l.id, p.id, NOW(), NOW()
FROM core.labels l, core.permissions p
WHERE l.key = 'calcutta_admin'
  AND l.deleted_at IS NULL
  AND p.key = 'entry.write'
  AND p.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM core.label_permissions lp
    WHERE lp.label_id = l.id AND lp.permission_id = p.id AND lp.deleted_at IS NULL
  );

-- Also link entry.write to site_admin label
INSERT INTO core.label_permissions (label_id, permission_id, created_at, updated_at)
SELECT l.id, p.id, NOW(), NOW()
FROM core.labels l, core.permissions p
WHERE l.key = 'site_admin'
  AND l.deleted_at IS NULL
  AND p.key = 'entry.write'
  AND p.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM core.label_permissions lp
    WHERE lp.label_id = l.id AND lp.permission_id = p.id AND lp.deleted_at IS NULL
  );
