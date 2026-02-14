-- Add predictions_json to lab.entries to separate market predictions from optimized bids
--
-- Pipeline stages:
-- 1. Model registered (lab.investment_models)
-- 2. Market predictions generated (predictions_json) - what model thinks market will bid
-- 3. Entry optimized (bids_json) - our optimal response to maximize ROI
-- 4. Entry evaluated (lab.evaluations) - Monte Carlo simulation results
--
-- predictions_json format:
-- [{
--   "team_id": "uuid",
--   "predicted_market_share": 0.045,  -- Model's prediction of team's share of pool
--   "expected_points": 12.5           -- Expected tournament points (from KenPom sim)
-- }]
--
-- bids_json format (existing, now represents OUR optimized bids):
-- [{
--   "team_id": "uuid",
--   "bid_points": 500,
--   "expected_roi": 2.5
-- }]

ALTER TABLE lab.entries
ADD COLUMN predictions_json JSONB;

COMMENT ON COLUMN lab.entries.predictions_json IS
    'Market predictions: [{team_id, predicted_market_share, expected_points}]. What model predicts others will bid.';

COMMENT ON COLUMN lab.entries.bids_json IS
    'Optimized bids: [{team_id, bid_points, expected_roi}]. Our optimal allocation given predictions.';

-- Update the model_leaderboard view to show pipeline progress
DROP VIEW IF EXISTS lab.model_leaderboard;
CREATE OR REPLACE VIEW lab.model_leaderboard AS
SELECT
    im.id AS investment_model_id,
    im.name AS model_name,
    im.kind AS model_kind,
    COUNT(DISTINCT e.id) AS n_entries,
    COUNT(DISTINCT CASE WHEN e.predictions_json IS NOT NULL THEN e.id END) AS n_entries_with_predictions,
    COUNT(ev.id) AS n_evaluations,
    COUNT(DISTINCT e.calcutta_id) AS n_calcuttas_with_entries,
    COUNT(DISTINCT CASE WHEN ev.id IS NOT NULL THEN e.calcutta_id END) AS n_calcuttas_with_evaluations,
    AVG(ev.mean_normalized_payout) AS avg_mean_payout,
    AVG(ev.median_normalized_payout) AS avg_median_payout,
    AVG(ev.p_top1) AS avg_p_top1,
    AVG(ev.p_in_money) AS avg_p_in_money,
    MIN(ev.created_at) AS first_eval_at,
    MAX(ev.created_at) AS last_eval_at
FROM lab.investment_models im
LEFT JOIN lab.entries e ON e.investment_model_id = im.id AND e.deleted_at IS NULL
LEFT JOIN lab.evaluations ev ON ev.entry_id = e.id AND ev.deleted_at IS NULL
WHERE im.deleted_at IS NULL
GROUP BY im.id, im.name, im.kind
ORDER BY avg_mean_payout DESC NULLS LAST;
