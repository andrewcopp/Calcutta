ALTER TABLE bronze_tournaments
    DROP CONSTRAINT IF EXISTS fk_bronze_tournaments_core_tournament_id;

DROP INDEX IF EXISTS idx_bronze_tournaments_core_tournament_id;

ALTER TABLE bronze_tournaments
    DROP COLUMN IF EXISTS core_tournament_id;
