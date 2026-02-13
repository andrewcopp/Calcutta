-- Restore original model_leaderboard view without calcutta count columns

DROP VIEW IF EXISTS lab.model_leaderboard;

CREATE VIEW lab.model_leaderboard AS
SELECT
    im.id AS investment_model_id,
    im.name AS model_name,
    im.kind AS model_kind,
    COUNT(DISTINCT e.id) AS n_entries,
    COUNT(ev.id) AS n_evaluations,
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
