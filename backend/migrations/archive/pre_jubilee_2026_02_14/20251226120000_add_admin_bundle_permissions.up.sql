INSERT INTO permissions (key, description)
VALUES
    ('admin.bundles.export', 'Export bundles'),
    ('admin.bundles.import', 'Import bundles'),
    ('admin.bundles.read', 'Read bundle upload status')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('admin.bundles.export', 'admin.bundles.import', 'admin.bundles.read')
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
