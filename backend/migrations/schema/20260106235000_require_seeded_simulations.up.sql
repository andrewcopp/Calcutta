UPDATE derived.simulated_tournaments
SET seed = 42
WHERE seed = 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_derived_simulated_tournaments_seed_nonzero'
    ) THEN
        ALTER TABLE derived.simulated_tournaments
            ADD CONSTRAINT ck_derived_simulated_tournaments_seed_nonzero
            CHECK (seed <> 0);
    END IF;
END $$;
