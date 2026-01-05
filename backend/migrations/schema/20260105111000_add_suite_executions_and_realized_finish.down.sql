ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    DROP COLUMN IF EXISTS suite_execution_id,
    DROP COLUMN IF EXISTS realized_finish_position,
    DROP COLUMN IF EXISTS realized_is_tied,
    DROP COLUMN IF EXISTS realized_in_the_money,
    DROP COLUMN IF EXISTS realized_payout_cents,
    DROP COLUMN IF EXISTS realized_total_points;

DROP TABLE IF EXISTS derived.suite_executions;
