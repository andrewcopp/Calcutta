DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('global_admin')
  AND p.key = 'admin.api_keys.write';

DELETE FROM permissions
WHERE key = 'admin.api_keys.write';

DROP TABLE IF EXISTS api_keys;
