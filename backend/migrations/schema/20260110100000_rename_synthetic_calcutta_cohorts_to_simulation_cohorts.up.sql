DO $$
BEGIN
    IF to_regclass('derived.synthetic_calcutta_cohorts') IS NOT NULL
       AND to_regclass('derived.simulation_cohorts') IS NULL THEN
        ALTER TABLE derived.synthetic_calcutta_cohorts RENAME TO simulation_cohorts;
    END IF;
END $$;
