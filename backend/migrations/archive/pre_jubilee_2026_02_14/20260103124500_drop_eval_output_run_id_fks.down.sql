-- Restore legacy FKs from analytics evaluation output rows to lab_gold.optimization_runs.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'analytics'
          AND t.relname = 'entry_simulation_outcomes'
          AND c.conname = 'gold_entry_simulation_outcomes_run_id_fkey'
    ) THEN
        ALTER TABLE analytics.entry_simulation_outcomes
            ADD CONSTRAINT gold_entry_simulation_outcomes_run_id_fkey
            FOREIGN KEY (run_id)
            REFERENCES lab_gold.optimization_runs(run_id)
            ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'analytics'
          AND t.relname = 'entry_performance'
          AND c.conname = 'gold_entry_performance_run_id_fkey'
    ) THEN
        ALTER TABLE analytics.entry_performance
            ADD CONSTRAINT gold_entry_performance_run_id_fkey
            FOREIGN KEY (run_id)
            REFERENCES lab_gold.optimization_runs(run_id)
            ON DELETE CASCADE;
    END IF;
END $$;
