package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	reb "github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultStrategyGenWorkerPollInterval = 2 * time.Second
	defaultStrategyGenWorkerStaleAfter   = 30 * time.Minute
)

type strategyGenJob struct {
	RunID     string
	RunKey    string
	Params    json.RawMessage
	ClaimedAt time.Time
}

func (s *Server) RunStrategyGenerationWorker(ctx context.Context) {
	s.RunStrategyGenerationWorkerWithOptions(ctx, defaultStrategyGenWorkerPollInterval, defaultStrategyGenWorkerStaleAfter)
}

func (s *Server) RunStrategyGenerationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("Strategy generation worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultStrategyGenWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultStrategyGenWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "strategy-generation-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			job, ok, err := s.claimNextStrategyGenerationJob(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next strategy generation job: %v", err)
				continue
			}
			if !ok {
				continue
			}
			_ = s.processStrategyGenerationJob(ctx, workerID, job)
		}
	}
}

func (s *Server) claimNextStrategyGenerationJob(ctx context.Context, workerID string, staleAfter time.Duration) (*strategyGenJob, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	staleBefore := now.Add(-staleAfter)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	job := &strategyGenJob{}
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'strategy_generation'
				AND (
					status = 'queued'
					OR (
						status = 'running'
						AND claimed_at IS NOT NULL
						AND claimed_at < $2
					)
				)
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE derived.run_jobs j
		SET status = 'running',
			attempt = j.attempt + 1,
			claimed_at = $1,
			claimed_by = $3,
			started_at = COALESCE(j.started_at, $1),
			finished_at = NULL,
			error_message = NULL,
			updated_at = NOW()
		FROM candidate
		WHERE j.id = candidate.id
		RETURNING j.run_id::text, j.run_key::text, j.params_json::text
	`

	var paramsStr string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{Time: staleBefore, Valid: true},
		workerID,
	).Scan(&job.RunID, &job.RunKey, &paramsStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	job.ClaimedAt = now
	job.Params = json.RawMessage([]byte(paramsStr))

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	committed = true

	return job, true, nil
}

func (s *Server) processStrategyGenerationJob(ctx context.Context, workerID string, job *strategyGenJob) bool {
	if job == nil {
		return false
	}

	log.Printf("strategy_generation_worker start worker_id=%s run_id=%s run_key=%s", workerID, job.RunID, job.RunKey)
	s.updateRunJobProgress(ctx, "strategy_generation", job.RunID, 0.05, "start", "Starting strategy generation job")
	s.updateRunJobProgress(ctx, "strategy_generation", job.RunID, 0.25, "running", "Generating recommended entry bids")

	svc := reb.New(s.pool)
	start := time.Now()
	res, err := svc.GenerateAndWriteToExistingStrategyGenerationRun(ctx, job.RunID)
	dur := time.Since(start)
	if err != nil {
		s.failStrategyGenerationJob(ctx, job, err)
		log.Printf("strategy_generation_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%v", workerID, job.RunID, job.RunKey, dur.Milliseconds(), err)
		return false
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'strategy_generation'
			AND run_id = $1::uuid
	`, job.RunID)

	s.updateRunJobProgress(ctx, "strategy_generation", job.RunID, 1.0, "succeeded", "Completed")

	var inputMarketShareArtifactID any
	if len(job.Params) > 0 {
		var params map[string]any
		if err := json.Unmarshal(job.Params, &params); err == nil {
			if v, ok := params["market_share_run_id"]; ok {
				if runIDStr, ok := v.(string); ok {
					runIDStr = strings.TrimSpace(runIDStr)
					if runIDStr != "" {
						var artifactID string
						_ = s.pool.QueryRow(ctx, `
							SELECT id::text
							FROM derived.run_artifacts
							WHERE run_kind = 'market_share'
								AND run_id = $1::uuid
								AND artifact_kind = 'metrics'
								AND deleted_at IS NULL
							LIMIT 1
					`, runIDStr).Scan(&artifactID)
						artifactID = strings.TrimSpace(artifactID)
						if artifactID != "" {
							inputMarketShareArtifactID = artifactID
						}
					}
				}
			}
		}
	}

	summary := map[string]any{
		"status":                  "succeeded",
		"strategyGenerationRunId": res.StrategyGenerationRunID,
		"runId":                   job.RunID,
		"runKey":                  job.RunKey,
		"nTeams":                  res.NTeams,
		"totalBidPoints":          res.TotalBidPoints,
		"simulatedTournamentId":   res.SimulatedTournamentID,
		"durationMs":              dur.Milliseconds(),
	}
	summaryJSON, jerr := json.Marshal(summary)
	if jerr == nil {
		var runKeyParam any
		if job.RunKey != "" {
			runKeyParam = job.RunKey
		} else {
			runKeyParam = nil
		}
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO derived.run_artifacts (
				run_kind,
				run_id,
				run_key,
				artifact_kind,
				schema_version,
				storage_uri,
				summary_json,
				input_market_share_artifact_id,
				input_advancement_artifact_id
			)
			VALUES ('strategy_generation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb, $4::uuid, NULL)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				input_market_share_artifact_id = EXCLUDED.input_market_share_artifact_id,
				input_advancement_artifact_id = EXCLUDED.input_advancement_artifact_id,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, summaryJSON, inputMarketShareArtifactID)
	}

	log.Printf("strategy_generation_worker success worker_id=%s run_id=%s run_key=%s n_teams=%d total_bid=%d dur_ms=%d",
		workerID,
		job.RunID,
		job.RunKey,
		res.NTeams,
		res.TotalBidPoints,
		dur.Milliseconds(),
	)
	return true
}

func (s *Server) failStrategyGenerationJob(ctx context.Context, job *strategyGenJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	if job != nil {
		s.updateRunJobProgress(ctx, "strategy_generation", job.RunID, 1.0, "failed", msg)
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'strategy_generation'
			AND run_id = $1::uuid
	`, job.RunID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"errorMessage": msg,
	}
	failureSummaryJSON, jerr := json.Marshal(failureSummary)
	if jerr == nil {
		var inputMarketShareArtifactID any
		if job != nil && len(job.Params) > 0 {
			var params map[string]any
			if err := json.Unmarshal(job.Params, &params); err == nil {
				if v, ok := params["market_share_run_id"]; ok {
					if runIDStr, ok := v.(string); ok {
						runIDStr = strings.TrimSpace(runIDStr)
						if runIDStr != "" {
							var artifactID string
							_ = s.pool.QueryRow(ctx, `
								SELECT id::text
								FROM derived.run_artifacts
								WHERE run_kind = 'market_share'
									AND run_id = $1::uuid
									AND artifact_kind = 'metrics'
									AND deleted_at IS NULL
								LIMIT 1
							`, runIDStr).Scan(&artifactID)
							artifactID = strings.TrimSpace(artifactID)
							if artifactID != "" {
								inputMarketShareArtifactID = artifactID
							}
						}
					}
				}
			}
		}

		var runKeyParam any
		if job.RunKey != "" {
			runKeyParam = job.RunKey
		} else {
			runKeyParam = nil
		}
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO derived.run_artifacts (
				run_kind,
				run_id,
				run_key,
				artifact_kind,
				schema_version,
				storage_uri,
				summary_json,
				input_market_share_artifact_id,
				input_advancement_artifact_id
			)
			VALUES ('strategy_generation', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb, $4::uuid, NULL)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				input_market_share_artifact_id = EXCLUDED.input_market_share_artifact_id,
				input_advancement_artifact_id = EXCLUDED.input_advancement_artifact_id,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, failureSummaryJSON, inputMarketShareArtifactID)
	}
}
