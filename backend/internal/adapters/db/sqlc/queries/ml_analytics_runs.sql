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

-- name: ListStrategyGenerationRunsByCoreCalcuttaID :many
SELECT
	sgr.id,
	sgr.run_key,
	sgr.simulated_tournament_id,
	sgr.calcutta_id,
	sgr.purpose,
	sgr.returns_model_key,
	sgr.investment_model_key,
	sgr.optimizer_key,
	sgr.params_json,
	sgr.git_sha,
	sgr.created_at
FROM derived.strategy_generation_runs sgr
WHERE sgr.calcutta_id = $1::uuid
	AND sgr.deleted_at IS NULL
ORDER BY sgr.created_at DESC;

-- name: GetStrategyGenerationRunByRunKey :one
SELECT
	sgr.id,
	COALESCE(sgr.run_key, ''::text) AS run_id,
	COALESCE(NULLIF(sgr.name, ''::text), COALESCE(sgr.run_key, ''::text)) AS name,
	sgr.calcutta_id,
	COALESCE(NULLIF(sgr.optimizer_key::text, ''::text), 'legacy'::text) AS strategy,
	COALESCE(tsb.n_sims, 0)::int AS n_sims,
	COALESCE(tsb.seed, 0)::int AS seed,
	COALESCE(c.budget_points, 100)::int AS budget_points,
	sgr.created_at
FROM derived.strategy_generation_runs sgr
LEFT JOIN core.calcuttas c
	ON c.id = sgr.calcutta_id
	AND c.deleted_at IS NULL
LEFT JOIN derived.simulated_tournaments tsb
	ON tsb.id = sgr.simulated_tournament_id
	AND tsb.deleted_at IS NULL
WHERE sgr.run_key = $1::text
	AND sgr.deleted_at IS NULL
LIMIT 1;
