-- Jubilee Baseline Down Migration
-- Drops all schemas to return to a clean state

DROP SCHEMA IF EXISTS lab CASCADE;
DROP SCHEMA IF EXISTS derived CASCADE;
DROP SCHEMA IF EXISTS core CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp";
