ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    ADD COLUMN IF NOT EXISTS optimizer_key TEXT,
    ADD COLUMN IF NOT EXISTS n_sims INT,
    ADD COLUMN IF NOT EXISTS seed INT,
    ADD COLUMN IF NOT EXISTS our_rank INT,
    ADD COLUMN IF NOT EXISTS our_mean_normalized_payout DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS our_median_normalized_payout DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS our_p_top1 DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS our_p_in_money DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS total_simulations INT;
