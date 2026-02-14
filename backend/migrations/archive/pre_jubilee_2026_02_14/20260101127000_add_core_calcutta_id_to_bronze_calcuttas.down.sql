DROP VIEW IF EXISTS bronze_calcuttas_core_ctx;

ALTER TABLE bronze_calcuttas
    DROP CONSTRAINT IF EXISTS fk_bronze_calcuttas_core_calcutta_id;

DROP INDEX IF EXISTS idx_bronze_calcuttas_core_calcutta_id;

ALTER TABLE bronze_calcuttas
    DROP COLUMN IF EXISTS core_calcutta_id;
