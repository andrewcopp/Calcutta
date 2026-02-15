DELETE FROM label_permissions lp
USING labels l, permissions p
WHERE lp.label_id = l.id
  AND lp.permission_id = p.id
  AND l.key IN ('global_admin')
  AND p.key IN ('admin.bundles.export', 'admin.bundles.import', 'admin.bundles.read');

DELETE FROM permissions
WHERE key IN ('admin.bundles.export', 'admin.bundles.import', 'admin.bundles.read');
