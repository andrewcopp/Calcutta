DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('global_admin')
  AND p.key IN (
    'analytics.suite_calcutta_evaluations.write',
    'analytics.suite_calcutta_evaluations.read'
  );

DELETE FROM permissions
WHERE key IN (
    'analytics.suite_calcutta_evaluations.write',
    'analytics.suite_calcutta_evaluations.read'
);

ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    DROP COLUMN IF EXISTS starting_state_key,
    DROP COLUMN IF EXISTS excluded_entry_name,
    DROP COLUMN IF EXISTS claimed_at,
    DROP COLUMN IF EXISTS claimed_by;
