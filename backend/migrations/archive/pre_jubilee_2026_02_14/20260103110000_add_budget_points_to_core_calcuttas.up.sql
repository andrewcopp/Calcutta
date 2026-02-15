ALTER TABLE core.calcuttas
    ADD COLUMN IF NOT EXISTS budget_points INTEGER NOT NULL DEFAULT 100;

CREATE INDEX IF NOT EXISTS idx_core_calcuttas_budget_points ON core.calcuttas(budget_points);
