-- Migration: remove_analytics_export_permission
-- Created: 2026-02-26 14:13:18 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Remove role_permission mapping for admin.analytics.export
DELETE FROM core.role_permissions
WHERE permission_id = '06f9f5af-0285-4852-9430-106aca75100c';

-- Remove the permission itself
DELETE FROM core.permissions
WHERE id = '06f9f5af-0285-4852-9430-106aca75100c';
