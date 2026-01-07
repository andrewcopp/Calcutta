package httpserver

import (
	"context"
	"encoding/json"
	"log"
)

type runProgressPayload struct {
	Percent float64 `json:"percent"`
	Phase   string  `json:"phase"`
	Message string  `json:"message"`
}

func (s *Server) updateRunJobProgress(ctx context.Context, runKind string, runID string, percent float64, phase string, message string) {
	if s == nil || s.pool == nil {
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
		return
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET progress_json = $3::jsonb,
			progress_updated_at = NOW(),
			updated_at = NOW()
		WHERE run_kind = $1
			AND run_id = $2::uuid
	`, runKind, runID, b)
	if err != nil {
		log.Printf("run_job_progress_update_failed run_kind=%s run_id=%s err=%v", runKind, runID, err)
	}
}
