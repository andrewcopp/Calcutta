-- Seed Data Down Migration
-- Removes permissions, labels, and label_permissions

SET search_path = '';

BEGIN;

TRUNCATE core.label_permissions CASCADE;
TRUNCATE core.labels CASCADE;
TRUNCATE core.permissions CASCADE;

COMMIT;
