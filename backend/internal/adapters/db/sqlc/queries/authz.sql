-- name: HasPermission :one
SELECT 1
FROM core.grants g
LEFT JOIN core.permissions p_direct ON g.permission_id = p_direct.id AND p_direct.deleted_at IS NULL
LEFT JOIN core.roles r ON g.role_id = r.id AND r.deleted_at IS NULL
LEFT JOIN core.role_permissions rp ON rp.role_id = r.id AND rp.deleted_at IS NULL
LEFT JOIN core.permissions p_role ON rp.permission_id = p_role.id AND p_role.deleted_at IS NULL
WHERE g.user_id = $1
  AND g.deleted_at IS NULL
  AND g.revoked_at IS NULL
  AND (g.expires_at IS NULL OR g.expires_at > NOW())
  AND (
    g.scope_type = 'global'
    OR (g.scope_type = $2 AND g.scope_id = NULLIF($3, '')::uuid)
  )
  AND ($4 = p_direct.key OR $4 = p_role.key)
LIMIT 1;
