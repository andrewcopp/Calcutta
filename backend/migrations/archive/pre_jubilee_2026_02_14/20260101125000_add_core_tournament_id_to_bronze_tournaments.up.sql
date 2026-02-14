ALTER TABLE bronze_tournaments
    ADD COLUMN IF NOT EXISTS core_tournament_id UUID;

UPDATE bronze_tournaments bt
SET core_tournament_id = (
    SELECT t.id
    FROM core.tournaments t
    JOIN core.seasons s ON s.id = t.season_id
    WHERE s.year = bt.season
      AND s.deleted_at IS NULL
      AND t.deleted_at IS NULL
    ORDER BY t.created_at DESC
    LIMIT 1
)
WHERE bt.core_tournament_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_bronze_tournaments_core_tournament_id
ON bronze_tournaments (core_tournament_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_bronze_tournaments_core_tournament_id'
    ) THEN
        ALTER TABLE bronze_tournaments
            ADD CONSTRAINT fk_bronze_tournaments_core_tournament_id
            FOREIGN KEY (core_tournament_id)
            REFERENCES core.tournaments(id)
            ON DELETE SET NULL;
    END IF;
END
$$;
