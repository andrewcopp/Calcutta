-- Make analytics.simulated_tournaments uniqueness batch-scoped so multiple
-- tournament simulation batches can coexist.

DO $$
DECLARE
    r record;
BEGIN
    -- Drop any UNIQUE constraints that enforce (tournament_id, sim_id, team_id).
    FOR r IN (
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'analytics.simulated_tournaments'::regclass
          AND contype = 'u'
          AND pg_get_constraintdef(oid) ILIKE '%tournament_id%'
          AND pg_get_constraintdef(oid) ILIKE '%sim_id%'
          AND pg_get_constraintdef(oid) ILIKE '%team_id%'
          AND pg_get_constraintdef(oid) NOT ILIKE '%tournament_simulation_batch_id%'
    ) LOOP
        EXECUTE format('ALTER TABLE analytics.simulated_tournaments DROP CONSTRAINT IF EXISTS %I', r.conname);
    END LOOP;

    -- Drop any UNIQUE indexes that enforce (tournament_id, sim_id, team_id).
    FOR r IN (
        SELECT indexname
        FROM pg_indexes
        WHERE schemaname = 'analytics'
          AND tablename = 'simulated_tournaments'
          AND indexdef ILIKE '%unique%'
          AND indexdef ILIKE '%(tournament_id, sim_id, team_id)%'
    ) LOOP
        EXECUTE format('DROP INDEX IF EXISTS analytics.%I', r.indexname);
    END LOOP;
END $$;

-- New uniqueness for lineage-native batches.
CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_simulated_tournaments_batch_sim_team
ON analytics.simulated_tournaments (tournament_simulation_batch_id, sim_id, team_id)
WHERE deleted_at IS NULL
  AND tournament_simulation_batch_id IS NOT NULL;

-- Preserve legacy uniqueness only for legacy rows (no batch id).
CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_simulated_tournaments_legacy_sim_team
ON analytics.simulated_tournaments (tournament_id, sim_id, team_id)
WHERE deleted_at IS NULL
  AND tournament_simulation_batch_id IS NULL;
