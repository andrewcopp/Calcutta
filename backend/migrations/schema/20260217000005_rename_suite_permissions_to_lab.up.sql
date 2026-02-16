-- Rename analytics.suites.* permissions to lab.* to match the new lab schema.

UPDATE core.permissions
SET key = 'lab.read',
    description = 'Read lab models, entries, evaluations'
WHERE id = 'eb55e249-6edd-42eb-8fc3-41b3920de2fa'
  AND deleted_at IS NULL;

UPDATE core.permissions
SET key = 'lab.write',
    description = 'Create/modify lab entries, pipelines'
WHERE id = '28df6d71-bf3e-44a8-b3aa-caa4a5be444c'
  AND deleted_at IS NULL;
