INSERT INTO permissions (key, description)
VALUES
    ('admin.analytics.read', 'Read analytics'),
    ('admin.hof.read', 'Read hall of fame')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN ('admin.analytics.read', 'admin.hof.read')
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;

DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('tournament_operator')
  AND p.key = 'admin.analytics.export';
