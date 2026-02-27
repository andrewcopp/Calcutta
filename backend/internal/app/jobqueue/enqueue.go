package jobqueue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Priority constants — lower value = higher priority.
const (
	PriorityCoreApp = 1
	PriorityLab     = 10
)

// Job kind constants.
const (
	KindRefreshPredictions = "refresh_predictions"
	KindRunSimulation      = "run_simulation"
	KindLabPredictions     = "lab_predictions"
	KindLabOptimization    = "lab_optimization"
	KindLabEvaluation      = "lab_evaluation"
)

// Enqueuer inserts jobs into the derived.run_jobs queue.
type Enqueuer struct {
	pool *pgxpool.Pool
}

// NewEnqueuer creates a new Enqueuer.
func NewEnqueuer(pool *pgxpool.Pool) *Enqueuer {
	return &Enqueuer{pool: pool}
}

// EnqueueResult contains the outcome of an enqueue attempt.
type EnqueueResult struct {
	JobID    string
	Enqueued bool // false when deduplicated (job already exists)
}

// Enqueue inserts a new job into the queue. If dedupKey is non-empty and a
// matching active job already exists, the insert is skipped (ON CONFLICT DO
// NOTHING) and Enqueued returns false.
func (e *Enqueuer) Enqueue(ctx context.Context, kind string, paramsJSON json.RawMessage, priority int, dedupKey string) (*EnqueueResult, error) {
	if paramsJSON == nil {
		paramsJSON = json.RawMessage(`{}`)
	}

	var dedupKeyPtr *string
	if dedupKey != "" {
		dedupKeyPtr = &dedupKey
	}

	var jobID string
	err := e.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status, priority, dedup_key)
		VALUES ($1, public.uuid_generate_v4(), public.uuid_generate_v4(), $2::jsonb, 'queued', $3, $4)
		ON CONFLICT (dedup_key) WHERE dedup_key IS NOT NULL AND status IN ('queued', 'running')
		DO NOTHING
		RETURNING run_id::text
	`, kind, paramsJSON, priority, dedupKeyPtr).Scan(&jobID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Deduplicated — job already exists
			return &EnqueueResult{Enqueued: false}, nil
		}
		return nil, err
	}
	return &EnqueueResult{JobID: jobID, Enqueued: true}, nil
}
