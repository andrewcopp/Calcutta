INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key = 'admin.analytics.export'
WHERE l.key IN ('tournament_operator')
ON CONFLICT DO NOTHING;

DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('global_admin')
  AND p.key IN ('admin.analytics.read', 'admin.hof.read');

DELETE FROM permissions
WHERE key IN ('admin.analytics.read', 'admin.hof.read');
