-- Drop triggers first
DROP TRIGGER IF EXISTS trg_derived_simulation_runs_enqueue_run_job ON derived.simulation_runs;
DROP TRIGGER IF EXISTS trg_derived_game_outcome_runs_enqueue_run_job ON derived.game_outcome_runs;
DROP TRIGGER IF EXISTS trg_derived_calcutta_evaluation_runs_enqueue_run_job ON derived.calcutta_evaluation_runs;

-- Drop trigger functions
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_simulation_run();
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_game_outcome_run();
DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_calcutta_evaluation_run();
