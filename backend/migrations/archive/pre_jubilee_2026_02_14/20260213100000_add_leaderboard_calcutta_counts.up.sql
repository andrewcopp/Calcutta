-- Add n_calcuttas_with_entries and n_calcuttas_with_evaluations to model_leaderboard view
-- These columns track how many calcuttas each model has covered, enabling progress visualization

DROP VIEW IF EXISTS lab.model_leaderboard;

CREATE VIEW lab.model_leaderboard AS
SELECT
    im.id AS investment_model_id,
    im.name AS model_name,
    im.kind AS model_kind,
    COUNT(DISTINCT e.id) AS n_entries,
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
