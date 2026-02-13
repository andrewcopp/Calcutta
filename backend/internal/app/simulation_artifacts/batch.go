package simulation_artifacts

import (
	"context"
)

func (s *Service) updateBatchStatus(ctx context.Context, simulationBatchID string) {
	if simulationBatchID == "" {
		return
	}
	_, _ = s.pool.Exec(ctx, `
		WITH agg AS (
			SELECT
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END)::int AS failed,
				SUM(CASE WHEN status IN ('queued', 'running') THEN 1 ELSE 0 END)::int AS pending
			FROM derived.simulation_runs
			WHERE simulation_run_batch_id = $1::uuid
				AND deleted_at IS NULL
		)
		UPDATE derived.simulation_run_batches e
		SET status = CASE
			WHEN a.failed > 0 THEN 'failed'
			WHEN a.pending > 0 THEN 'running'
			ELSE 'succeeded'
		END,
			error_message = CASE
			WHEN a.failed > 0 THEN COALESCE((
				SELECT error_message
				FROM derived.simulation_runs
				WHERE simulation_run_batch_id = $1::uuid
					AND status = 'failed'
					AND error_message IS NOT NULL
					AND deleted_at IS NULL
				LIMIT 1
			), e.error_message)
			ELSE NULL
		END,
			updated_at = NOW()
		FROM agg a
		WHERE e.id = $1::uuid
			AND e.deleted_at IS NULL
	`, simulationBatchID)
}
