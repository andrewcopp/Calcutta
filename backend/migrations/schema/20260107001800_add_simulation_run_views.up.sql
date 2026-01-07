CREATE SCHEMA IF NOT EXISTS derived;

CREATE OR REPLACE VIEW derived.simulation_run_batches AS
SELECT
    e.id,
    e.suite_id,
    e.suite_id AS cohort_id,
    e.name,
    e.optimizer_key,
    e.n_sims,
    e.seed,
    e.starting_state_key,
    e.excluded_entry_name,
    e.status,
    e.error_message,
    e.created_at,
    e.updated_at,
    e.deleted_at
FROM derived.suite_executions e;

CREATE OR REPLACE VIEW derived.simulation_runs AS
SELECT
    r.id,
    r.run_key,
    r.suite_execution_id,
    r.suite_execution_id AS simulation_run_batch_id,
    r.suite_id,
    r.suite_id AS cohort_id,
    r.calcutta_id,
    r.game_outcome_run_id,
    r.market_share_run_id,
    r.strategy_generation_run_id,
    r.calcutta_evaluation_run_id,
    r.starting_state_key,
    r.excluded_entry_name,
    r.optimizer_key,
    r.n_sims,
    r.seed,
    r.our_rank,
    r.our_mean_normalized_payout,
    r.our_median_normalized_payout,
    r.our_p_top1,
    r.our_p_in_money,
    r.total_simulations,
    r.realized_finish_position,
    r.realized_is_tied,
    r.realized_in_the_money,
    r.realized_payout_cents,
    r.realized_total_points,
    r.status,
    r.claimed_at,
    r.claimed_by,
    r.error_message,
    r.created_at,
    r.updated_at,
    r.deleted_at
FROM derived.suite_calcutta_evaluations r;
