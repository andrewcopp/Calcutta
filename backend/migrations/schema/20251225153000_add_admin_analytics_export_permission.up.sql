INSERT INTO permissions (key, description)
VALUES ('admin.analytics.export', 'Export analytics snapshots')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key = 'admin.analytics.export'
WHERE l.key IN ('global_admin', 'tournament_operator')
ON CONFLICT DO NOTHING;
