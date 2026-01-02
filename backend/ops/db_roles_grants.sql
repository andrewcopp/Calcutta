DO $$
BEGIN
    CREATE ROLE app_writer;
EXCEPTION
    WHEN duplicate_object THEN
        NULL;
END
$$;

DO $$
BEGIN
    CREATE ROLE lab_pipeline;
EXCEPTION
    WHEN duplicate_object THEN
        NULL;
END
$$;

-- Core schema privileges
GRANT USAGE ON SCHEMA core TO app_writer;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA core TO app_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA core GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_writer;

GRANT USAGE ON SCHEMA core TO lab_pipeline;
GRANT SELECT ON ALL TABLES IN SCHEMA core TO lab_pipeline;
ALTER DEFAULT PRIVILEGES IN SCHEMA core GRANT SELECT ON TABLES TO lab_pipeline;

-- Medallion schema privileges
GRANT USAGE ON SCHEMA bronze TO lab_pipeline;
GRANT CREATE ON SCHEMA bronze TO lab_pipeline;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA bronze TO lab_pipeline;
ALTER DEFAULT PRIVILEGES IN SCHEMA bronze GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lab_pipeline;

GRANT USAGE ON SCHEMA silver TO lab_pipeline;
GRANT CREATE ON SCHEMA silver TO lab_pipeline;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA silver TO lab_pipeline;
ALTER DEFAULT PRIVILEGES IN SCHEMA silver GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lab_pipeline;

GRANT USAGE ON SCHEMA gold TO lab_pipeline;
GRANT CREATE ON SCHEMA gold TO lab_pipeline;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA gold TO lab_pipeline;
ALTER DEFAULT PRIVILEGES IN SCHEMA gold GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lab_pipeline;

-- Optional: allow app to read lab tables
GRANT USAGE ON SCHEMA bronze TO app_writer;
GRANT USAGE ON SCHEMA silver TO app_writer;
GRANT USAGE ON SCHEMA gold TO app_writer;
GRANT SELECT ON ALL TABLES IN SCHEMA bronze TO app_writer;
GRANT SELECT ON ALL TABLES IN SCHEMA silver TO app_writer;
GRANT SELECT ON ALL TABLES IN SCHEMA gold TO app_writer;

-- Optional: grant roles to current DB user (adjust as needed)
DO $$
BEGIN
    GRANT app_writer TO calcutta;
EXCEPTION
    WHEN undefined_object THEN
        NULL;
END
$$;

DO $$
BEGIN
    GRANT lab_pipeline TO calcutta;
EXCEPTION
    WHEN undefined_object THEN
        NULL;
END
$$;
