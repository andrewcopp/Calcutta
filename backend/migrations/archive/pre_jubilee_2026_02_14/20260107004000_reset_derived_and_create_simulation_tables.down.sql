DROP TRIGGER IF EXISTS trg_derived_simulation_runs_enqueue_run_job ON derived.simulation_runs;
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_simulation_run();

DROP TABLE IF EXISTS derived.simulation_runs;
DROP TABLE IF EXISTS derived.simulation_run_batches;
DROP TABLE IF EXISTS derived.synthetic_calcuttas;
DROP TABLE IF EXISTS derived.synthetic_calcutta_cohorts;

DROP TABLE IF EXISTS derived.run_artifacts;
DROP TABLE IF EXISTS derived.run_jobs;
