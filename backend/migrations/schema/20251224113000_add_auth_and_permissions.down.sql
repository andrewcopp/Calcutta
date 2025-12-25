DROP TABLE IF EXISTS grants;
DROP TABLE IF EXISTS label_permissions;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS auth_sessions;

ALTER TABLE users
    DROP COLUMN IF EXISTS password_hash;
