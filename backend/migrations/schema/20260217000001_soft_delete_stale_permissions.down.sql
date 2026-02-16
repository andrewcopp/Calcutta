-- Restore soft-deleted permissions
UPDATE core.label_permissions
SET deleted_at = NULL
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
);

UPDATE core.permissions
SET deleted_at = NULL
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
);
