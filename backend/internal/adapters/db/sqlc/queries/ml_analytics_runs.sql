-- name: ListCalcuttaEvaluationRunsByCoreCalcuttaID :many
SELECT
	cer.id,
	cer.simulated_tournament_id,
	cer.calcutta_snapshot_id,
	cer.purpose,
	cer.created_at
FROM derived.calcutta_evaluation_runs cer
JOIN core.calcutta_snapshots cs
	ON cs.id = cer.calcutta_snapshot_id
	AND cs.deleted_at IS NULL
WHERE cs.base_calcutta_id = $1::uuid
	AND cer.deleted_at IS NULL
ORDER BY cer.created_at DESC;

-- name: ListOptimizedEntriesByCoreCalcuttaID :many
SELECT
	oe.id,
	oe.run_key,
	oe.simulated_tournament_id,
	oe.calcutta_id,
	oe.purpose,
	oe.returns_model_key,
	oe.investment_model_key,
	oe.optimizer_kind,
	oe.optimizer_params_json AS params_json,
	oe.git_sha,
	oe.created_at
FROM derived.optimized_entries oe
WHERE oe.calcutta_id = $1::uuid
	AND oe.deleted_at IS NULL
ORDER BY oe.created_at DESC;

-- name: GetOptimizedEntryByRunKey :one
SELECT
	oe.id,
	COALESCE(oe.run_key, ''::text) AS run_id,
	COALESCE(NULLIF(oe.name, ''::text), COALESCE(oe.run_key, ''::text)) AS name,
	oe.calcutta_id,
	COALESCE(NULLIF(oe.optimizer_kind::text, ''::text), 'legacy'::text) AS strategy,
	COALESCE(tsb.n_sims, 0)::int AS n_sims,
	COALESCE(tsb.seed, 0)::int AS seed,
	COALESCE(c.budget_points, 100)::int AS budget_points,
	oe.created_at
FROM derived.optimized_entries oe
LEFT JOIN core.calcuttas c
	ON c.id = oe.calcutta_id
	AND c.deleted_at IS NULL
LEFT JOIN derived.simulated_tournaments tsb
	ON tsb.id = oe.simulated_tournament_id
	AND tsb.deleted_at IS NULL
WHERE oe.run_key = $1::text
	AND oe.deleted_at IS NULL
LIMIT 1;
