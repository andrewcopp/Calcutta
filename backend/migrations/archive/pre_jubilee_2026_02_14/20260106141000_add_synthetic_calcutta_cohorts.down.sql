DROP TRIGGER IF EXISTS trg_sync_synthetic_calcutta_cohort_from_suite ON derived.suites;
DROP FUNCTION IF EXISTS derived.sync_synthetic_calcutta_cohort_from_suite();
DROP TABLE IF EXISTS derived.synthetic_calcutta_cohorts;
