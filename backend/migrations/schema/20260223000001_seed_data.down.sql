-- Seed Data Down Migration
-- Removes permissions, roles, and role_permissions

SET search_path = '';

BEGIN;

TRUNCATE core.role_permissions CASCADE;
TRUNCATE core.roles CASCADE;
TRUNCATE core.permissions CASCADE;

COMMIT;
