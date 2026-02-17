-- lab.entry_evaluations: view that joined lab.entries with lab.evaluations, never read by app code
DROP VIEW IF EXISTS lab.entry_evaluations;

-- derived.game_outcome_runs.prediction_model_id: FK to prediction_models, never used in app code
ALTER TABLE IF EXISTS derived.game_outcome_runs
    DROP CONSTRAINT IF EXISTS game_outcome_runs_prediction_model_id_fkey;
ALTER TABLE IF EXISTS derived.game_outcome_runs
    DROP COLUMN IF EXISTS prediction_model_id;

-- derived.entry_performance: only written by deleted writer.go, never read
DROP TABLE IF EXISTS derived.entry_performance CASCADE;

-- derived.calcutta_evaluation_runs: only written by deleted createCalcuttaEvaluationRun, never read
DROP TABLE IF EXISTS derived.calcutta_evaluation_runs CASCADE;

-- derived.prediction_models: only referenced by dropped FK above, never used in app code
DROP TABLE IF EXISTS derived.prediction_models CASCADE;
