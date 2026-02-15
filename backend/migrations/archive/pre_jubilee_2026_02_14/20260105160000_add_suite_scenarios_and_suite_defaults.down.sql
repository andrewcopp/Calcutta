DROP TRIGGER IF EXISTS set_updated_at ON derived.suite_scenarios;
DROP TABLE IF EXISTS derived.suite_scenarios CASCADE;

DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('global_admin')
  AND p.key IN (
    'analytics.suites.write',
    'analytics.suites.read',
    'analytics.suite_scenarios.write',
    'analytics.suite_scenarios.read',
    'analytics.strategy_generation_runs.write',
    'analytics.strategy_generation_runs.read'
  );

DELETE FROM permissions
WHERE key IN (
  'analytics.suites.write',
  'analytics.suites.read',
  'analytics.suite_scenarios.write',
  'analytics.suite_scenarios.read',
  'analytics.strategy_generation_runs.write',
  'analytics.strategy_generation_runs.read'
);

ALTER TABLE IF EXISTS derived.suites
    DROP CONSTRAINT IF EXISTS ck_derived_suites_starting_state_key;

ALTER TABLE IF EXISTS derived.suites
    DROP COLUMN IF EXISTS excluded_entry_name;

ALTER TABLE IF EXISTS derived.suites
    DROP COLUMN IF EXISTS starting_state_key;
