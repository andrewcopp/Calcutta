-- Verification checks for backend/ops/db_roles_grants.sql
-- Intended to be run with psql -v ON_ERROR_STOP=1

-- 1) Roles exist
SELECT 'role_exists_app_writer' AS check, EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_writer') AS ok;
SELECT 'role_exists_lab_pipeline' AS check, EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'lab_pipeline') AS ok;

-- 2) Schema privileges
SELECT 'app_writer_core_usage' AS check, has_schema_privilege('app_writer', 'core', 'USAGE') AS ok;
SELECT 'lab_pipeline_core_usage' AS check, has_schema_privilege('lab_pipeline', 'core', 'USAGE') AS ok;

SELECT 'app_writer_bronze_usage' AS check, has_schema_privilege('app_writer', 'bronze', 'USAGE') AS ok;
SELECT 'app_writer_silver_usage' AS check, has_schema_privilege('app_writer', 'silver', 'USAGE') AS ok;
SELECT 'app_writer_gold_usage' AS check, has_schema_privilege('app_writer', 'gold', 'USAGE') AS ok;

SELECT 'lab_pipeline_bronze_create' AS check, has_schema_privilege('lab_pipeline', 'bronze', 'CREATE') AS ok;
SELECT 'lab_pipeline_silver_create' AS check, has_schema_privilege('lab_pipeline', 'silver', 'CREATE') AS ok;
SELECT 'lab_pipeline_gold_create' AS check, has_schema_privilege('lab_pipeline', 'gold', 'CREATE') AS ok;

-- 3) Spot-check table privileges
SELECT 'app_writer_core_tournaments_select' AS check, has_table_privilege('app_writer', 'core.tournaments', 'SELECT') AS ok;
SELECT 'app_writer_core_tournaments_insert' AS check, has_table_privilege('app_writer', 'core.tournaments', 'INSERT') AS ok;
SELECT 'app_writer_core_tournaments_update' AS check, has_table_privilege('app_writer', 'core.tournaments', 'UPDATE') AS ok;
SELECT 'app_writer_core_tournaments_delete' AS check, has_table_privilege('app_writer', 'core.tournaments', 'DELETE') AS ok;

SELECT 'lab_pipeline_core_tournaments_select' AS check, has_table_privilege('lab_pipeline', 'core.tournaments', 'SELECT') AS ok;

-- 4) Membership: calcutta user should have app_writer and lab_pipeline granted (if user exists)
SELECT 'calcutta_has_app_writer' AS check,
       CASE WHEN EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'calcutta')
            THEN pg_has_role('calcutta', 'app_writer', 'member')
            ELSE NULL
       END AS ok;

SELECT 'calcutta_has_lab_pipeline' AS check,
       CASE WHEN EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'calcutta')
            THEN pg_has_role('calcutta', 'lab_pipeline', 'member')
            ELSE NULL
       END AS ok;
