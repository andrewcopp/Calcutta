-- Drop legacy FKs that force analytics evaluation output rows to reference lab_gold.optimization_runs.
-- Evaluation outputs should be keyed by calcutta_evaluation_run_id; run_id is retained for legacy
-- but should not block inserts.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'analytics'
          AND t.relname = 'entry_simulation_outcomes'
          AND c.conname = 'gold_entry_simulation_outcomes_run_id_fkey'
    ) THEN
        ALTER TABLE analytics.entry_simulation_outcomes
            DROP CONSTRAINT gold_entry_simulation_outcomes_run_id_fkey;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'analytics'
          AND t.relname = 'entry_performance'
          AND c.conname = 'gold_entry_performance_run_id_fkey'
    ) THEN
        ALTER TABLE analytics.entry_performance
            DROP CONSTRAINT gold_entry_performance_run_id_fkey;
    END IF;
END $$;
