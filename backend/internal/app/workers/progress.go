package workers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgressWriter interface {
	Update(ctx context.Context, runKind string, runID string, percent float64, phase string, message string)
}

type DBProgressWriter struct {
	pool *pgxpool.Pool
}

func NewDBProgressWriter(pool *pgxpool.Pool) *DBProgressWriter {
	return &DBProgressWriter{pool: pool}
}

type runProgressPayload struct {
	Percent float64 `json:"percent"`
	Phase   string  `json:"phase"`
	Message string  `json:"message"`
}

func (w *DBProgressWriter) Update(ctx context.Context, runKind string, runID string, percent float64, phase string, message string) {
	if w == nil || w.pool == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	payload := runProgressPayload{Percent: percent, Phase: phase, Message: message}
	b, err := json.Marshal(payload)
	if err != nil {
		slog.Error("run_job_progress_marshal_failed", "run_kind", runKind, "run_id", runID, "error", err)
		return
	}

	_, err = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET progress_json = $3::jsonb,
			progress_updated_at = NOW(),
			updated_at = NOW()
		WHERE run_kind = $1
			AND run_id = $2::uuid
	`, runKind, runID, b)
	if err != nil {
		slog.Error("run_job_progress_update_failed", "run_kind", runKind, "run_id", runID, "error", err)
	}
}
