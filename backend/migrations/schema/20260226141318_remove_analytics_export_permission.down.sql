-- Rollback: remove_analytics_export_permission
-- Created: 2026-02-26 14:13:18 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Re-insert the permission
INSERT INTO core.permissions (id, name, description)
VALUES ('06f9f5af-0285-4852-9430-106aca75100c', 'admin.analytics.export', 'Export analytics snapshots');

-- Re-insert the role_permission mapping (admin role)
INSERT INTO core.role_permissions (role_id, permission_id)
VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '06f9f5af-0285-4852-9430-106aca75100c');
