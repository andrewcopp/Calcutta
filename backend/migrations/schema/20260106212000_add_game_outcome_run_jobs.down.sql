DELETE FROM derived.run_jobs
WHERE run_kind = 'game_outcome';

DROP TRIGGER IF EXISTS trg_derived_game_outcome_runs_enqueue_run_job ON derived.game_outcome_runs;

DROP FUNCTION IF EXISTS derived.enqueue_run_job_for_game_outcome_run();

DROP INDEX IF EXISTS idx_derived_game_outcome_runs_run_key;

ALTER TABLE IF EXISTS derived.game_outcome_runs
    DROP COLUMN IF EXISTS run_key;
