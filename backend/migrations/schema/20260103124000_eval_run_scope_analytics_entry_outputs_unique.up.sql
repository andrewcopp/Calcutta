-- Make analytics.entry_simulation_outcomes and analytics.entry_performance uniqueness
-- calcutta_evaluation_run scoped so multiple eval runs can coexist.

DO $$
DECLARE
    r record;
BEGIN
    -- Drop UNIQUE constraints on entry_simulation_outcomes that don't include calcutta_evaluation_run_id.
    FOR r IN (
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'analytics.entry_simulation_outcomes'::regclass
          AND contype = 'u'
          AND pg_get_constraintdef(oid) NOT ILIKE '%calcutta_evaluation_run_id%'
    ) LOOP
        EXECUTE format('ALTER TABLE analytics.entry_simulation_outcomes DROP CONSTRAINT IF EXISTS %I', r.conname);
    END LOOP;

    -- Drop UNIQUE indexes on entry_simulation_outcomes that don't include calcutta_evaluation_run_id.
    FOR r IN (
        SELECT indexname
        FROM pg_indexes
        WHERE schemaname = 'analytics'
          AND tablename = 'entry_simulation_outcomes'
          AND indexdef ILIKE '%unique%'
          AND indexdef NOT ILIKE '%calcutta_evaluation_run_id%'
          AND indexname NOT LIKE '%_pkey'
    ) LOOP
        EXECUTE format('DROP INDEX IF EXISTS analytics.%I', r.indexname);
    END LOOP;

    -- Drop UNIQUE constraints on entry_performance that don't include calcutta_evaluation_run_id.
    FOR r IN (
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'analytics.entry_performance'::regclass
          AND contype = 'u'
          AND pg_get_constraintdef(oid) NOT ILIKE '%calcutta_evaluation_run_id%'
    ) LOOP
        EXECUTE format('ALTER TABLE analytics.entry_performance DROP CONSTRAINT IF EXISTS %I', r.conname);
    END LOOP;

    -- Drop UNIQUE indexes on entry_performance that don't include calcutta_evaluation_run_id.
    FOR r IN (
        SELECT indexname
        FROM pg_indexes
        WHERE schemaname = 'analytics'
          AND tablename = 'entry_performance'
          AND indexdef ILIKE '%unique%'
          AND indexdef NOT ILIKE '%calcutta_evaluation_run_id%'
          AND indexname NOT LIKE '%_pkey'
    ) LOOP
        EXECUTE format('DROP INDEX IF EXISTS analytics.%I', r.indexname);
    END LOOP;
END $$;

-- Eval-run scoped uniqueness.
CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_sim_outcomes_eval_run_entry_sim
ON analytics.entry_simulation_outcomes (calcutta_evaluation_run_id, entry_name, sim_id)
WHERE deleted_at IS NULL
  AND calcutta_evaluation_run_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_performance_eval_run_entry
ON analytics.entry_performance (calcutta_evaluation_run_id, entry_name)
WHERE deleted_at IS NULL
  AND calcutta_evaluation_run_id IS NOT NULL;

-- Legacy uniqueness for legacy rows.
CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_sim_outcomes_legacy_run_entry_sim
ON analytics.entry_simulation_outcomes (run_id, entry_name, sim_id)
WHERE deleted_at IS NULL
  AND calcutta_evaluation_run_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_entry_performance_legacy_run_entry
ON analytics.entry_performance (run_id, entry_name)
WHERE deleted_at IS NULL
  AND calcutta_evaluation_run_id IS NULL;
