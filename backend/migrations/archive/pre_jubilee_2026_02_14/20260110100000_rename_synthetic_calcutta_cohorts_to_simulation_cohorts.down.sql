DO $$
BEGIN
    IF to_regclass('derived.simulation_cohorts') IS NOT NULL
       AND to_regclass('derived.synthetic_calcutta_cohorts') IS NULL THEN
        ALTER TABLE derived.simulation_cohorts RENAME TO synthetic_calcutta_cohorts;
    END IF;
END $$;
