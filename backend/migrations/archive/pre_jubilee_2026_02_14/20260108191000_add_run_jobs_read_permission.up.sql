INSERT INTO permissions (key, description)
VALUES
    ('analytics.run_jobs.read', 'Read run job progress and status')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('analytics.run_jobs.read')
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
