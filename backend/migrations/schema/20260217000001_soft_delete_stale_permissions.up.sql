-- Soft-delete permission keys for removed analytics tables.
-- These permissions referenced tables that no longer exist (suite_*, strategy_generation_runs, etc.)

UPDATE core.permissions
SET deleted_at = NOW()
WHERE key IN (
    'analytics.entry_evaluation_requests.write',
    'analytics.entry_evaluation_requests.read',
    'analytics.suite_calcutta_evaluations.write',
    'analytics.suite_calcutta_evaluations.read',
    'analytics.suite_executions.write',
    'analytics.suite_executions.read',
    'analytics.suite_scenarios.write',
    'analytics.suite_scenarios.read',
    'analytics.strategy_generation_runs.write',
    'analytics.strategy_generation_runs.read',
    'analytics.run_jobs.read'
)
AND deleted_at IS NULL;

-- Also soft-delete any label_permissions rows referencing these permissions
UPDATE core.label_permissions
SET deleted_at = NOW()
WHERE permission_id IN (
    SELECT id FROM core.permissions
    WHERE key IN (
        'analytics.entry_evaluation_requests.write',
        'analytics.entry_evaluation_requests.read',
        'analytics.suite_calcutta_evaluations.write',
        'analytics.suite_calcutta_evaluations.read',
        'analytics.suite_executions.write',
        'analytics.suite_executions.read',
        'analytics.suite_scenarios.write',
        'analytics.suite_scenarios.read',
        'analytics.strategy_generation_runs.write',
        'analytics.strategy_generation_runs.read',
        'analytics.run_jobs.read'
    )
)
AND deleted_at IS NULL;
