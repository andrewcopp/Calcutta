-- Revert lab.* permissions back to analytics.suites.* names.

UPDATE core.permissions
SET key = 'analytics.suites.read',
    description = 'Read suites'
WHERE id = 'eb55e249-6edd-42eb-8fc3-41b3920de2fa'
  AND deleted_at IS NULL;

UPDATE core.permissions
SET key = 'analytics.suites.write',
    description = 'Create/update suites'
WHERE id = '28df6d71-bf3e-44a8-b3aa-caa4a5be444c'
  AND deleted_at IS NULL;
