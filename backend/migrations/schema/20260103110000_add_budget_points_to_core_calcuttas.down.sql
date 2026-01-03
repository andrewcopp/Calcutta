ALTER TABLE core.calcuttas
    DROP COLUMN IF EXISTS budget_points;

DROP INDEX IF EXISTS idx_core_calcuttas_budget_points;
