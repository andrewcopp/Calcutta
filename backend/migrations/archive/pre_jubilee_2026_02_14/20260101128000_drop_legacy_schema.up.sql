-- Drops the legacy schema after the verification window.
-- This is intentionally destructive and should only be run once the app and tooling are fully cut over to core/bronze/silver/gold.

DROP SCHEMA IF EXISTS legacy CASCADE;
