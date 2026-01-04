INSERT INTO permissions (key, description)
VALUES
    ('analytics.entry_evaluation_requests.read', 'Read entry evaluation requests')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('analytics.entry_evaluation_requests.read')
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
