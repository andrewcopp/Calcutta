ALTER TABLE IF EXISTS derived.simulated_tournaments
    DROP CONSTRAINT IF EXISTS ck_derived_simulated_tournaments_seed_nonzero;
