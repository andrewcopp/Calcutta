ALTER TABLE bronze_calcuttas
    ADD COLUMN IF NOT EXISTS core_calcutta_id UUID;

UPDATE bronze_calcuttas bc
SET core_calcutta_id = (
    SELECT c.id
    FROM core.calcuttas c
    JOIN bronze_tournaments bt ON bt.id = bc.tournament_id
    WHERE c.tournament_id = bt.core_tournament_id
      AND c.name = bc.name
      AND c.deleted_at IS NULL
    ORDER BY c.created_at DESC
    LIMIT 1
)
WHERE bc.core_calcutta_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_bronze_calcuttas_core_calcutta_id
ON bronze_calcuttas (core_calcutta_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_bronze_calcuttas_core_calcutta_id'
    ) THEN
        ALTER TABLE bronze_calcuttas
            ADD CONSTRAINT fk_bronze_calcuttas_core_calcutta_id
            FOREIGN KEY (core_calcutta_id)
            REFERENCES core.calcuttas(id)
            ON DELETE SET NULL;
    END IF;
END
$$;

CREATE OR REPLACE VIEW bronze_calcuttas_core_ctx AS
SELECT
    bc.id,
    bc.tournament_id,
    bt.season,
    bt.core_tournament_id,
    bc.core_calcutta_id,
    bc.name,
    bc.min_teams,
    bc.max_teams,
    bc.max_bid_points,
    bc.created_at
FROM bronze_calcuttas bc
JOIN bronze_tournaments bt ON bt.id = bc.tournament_id;
