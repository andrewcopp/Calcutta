INSERT INTO permissions (key, description)
VALUES
    ('admin.users.read', 'Read users')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('admin.users.read')
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
