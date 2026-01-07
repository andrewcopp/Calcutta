package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultMarketShareWorkerPollInterval = 2 * time.Second
	defaultMarketShareWorkerStaleAfter   = 30 * time.Minute
)

type marketShareJob struct {
	RunID     string
	RunKey    string
	Params    json.RawMessage
	ClaimedAt time.Time
}

type pythonMarketShareRunnerResult struct {
	OK           bool    `json:"ok"`
	RunID        *string `json:"run_id"`
	RowsInserted *int    `json:"rows_inserted"`
	Error        *string `json:"error"`
}

func (s *Server) RunMarketShareWorker(ctx context.Context) {
	s.RunMarketShareWorkerWithOptions(ctx, defaultMarketShareWorkerPollInterval, defaultMarketShareWorkerStaleAfter)
}

func (s *Server) RunMarketShareWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("Market share worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultMarketShareWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultMarketShareWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "market-share-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			job, ok, err := s.claimNextMarketShareJob(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next market share job: %v", err)
				continue
			}
			if !ok {
				continue
			}
			_ = s.processMarketShareJob(ctx, workerID, job)
		}
	}
}

func (s *Server) claimNextMarketShareJob(ctx context.Context, workerID string, staleAfter time.Duration) (*marketShareJob, bool, error) {
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

	job := &marketShareJob{}
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind = 'market_share'
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

func (s *Server) processMarketShareJob(ctx context.Context, workerID string, job *marketShareJob) bool {
	if job == nil {
		return false
	}

	pythonBin := strings.TrimSpace(os.Getenv("PYTHON_BIN"))
	if pythonBin == "" {
		pythonBin = "python3"
	}

	runnerPath := strings.TrimSpace(os.Getenv("PYTHON_MARKET_SHARE_RUNNER"))
	candidates := make([]string, 0, 2)
	if runnerPath != "" {
		candidates = append(candidates, runnerPath)
	}
	candidates = append(candidates,
		"data-science/scripts/run_market_share_runner.py",
		"../data-science/scripts/run_market_share_runner.py",
	)

	resolvedRunner := ""
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			resolvedRunner = abs
			break
		}
	}
	if resolvedRunner == "" {
		err := errors.New("market share python runner not found; set PYTHON_MARKET_SHARE_RUNNER")
		s.failMarketShareJob(ctx, job, err)
		log.Printf("market_share_worker fail worker_id=%s run_id=%s run_key=%s err=%v", workerID, job.RunID, job.RunKey, err)
		return false
	}

	log.Printf("market_share_worker start worker_id=%s run_id=%s run_key=%s", workerID, job.RunID, job.RunKey)

	cmd := exec.CommandContext(
		ctx,
		pythonBin,
		resolvedRunner,
		"--run-id",
		job.RunID,
	)
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)
	outStr := strings.TrimSpace(stdout.String())
	if outStr == "" {
		outStr = "{}"
	}

	var parsed pythonMarketShareRunnerResult
	_ = json.Unmarshal([]byte(outStr), &parsed)

	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			msg = *parsed.Error
		}
		if msg == "" {
			msg = err.Error()
		}
		s.failMarketShareJob(ctx, job, errors.New(msg))
		log.Printf("market_share_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%s", workerID, job.RunID, job.RunKey, dur.Milliseconds(), msg)
		return false
	}
	if !parsed.OK {
		msg := "python runner returned ok=false"
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			msg = *parsed.Error
		}
		s.failMarketShareJob(ctx, job, errors.New(msg))
		log.Printf("market_share_worker fail worker_id=%s run_id=%s run_key=%s dur_ms=%d err=%s", workerID, job.RunID, job.RunKey, dur.Milliseconds(), msg)
		return false
	}

	rowsInserted := 0
	if parsed.RowsInserted != nil {
		rowsInserted = *parsed.RowsInserted
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded',
			finished_at = NOW(),
			error_message = NULL,
			updated_at = NOW()
		WHERE run_kind = 'market_share'
			AND run_id = $1::uuid
	`, job.RunID)

	summary := map[string]any{
		"status":       "succeeded",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"rowsInserted": rowsInserted,
		"durationMs":   dur.Milliseconds(),
	}
	if len(job.Params) > 0 {
		var p any
		if err := json.Unmarshal(job.Params, &p); err == nil {
			summary["params"] = p
		}
	}
	summaryJSON, jerr := json.Marshal(summary)
	if jerr == nil {
		var runKeyParam any
		if strings.TrimSpace(job.RunKey) != "" {
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
				summary_json
			)
			VALUES ('market_share', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, summaryJSON)
	}

	log.Printf("market_share_worker success worker_id=%s run_id=%s run_key=%s rows_inserted=%d dur_ms=%d", workerID, job.RunID, job.RunKey, rowsInserted, dur.Milliseconds())
	return true
}

func (s *Server) failMarketShareJob(ctx context.Context, job *marketShareJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = $2,
			updated_at = NOW()
		WHERE run_kind = 'market_share'
			AND run_id = $1::uuid
	`, job.RunID, msg)

	failureSummary := map[string]any{
		"status":       "failed",
		"runId":        job.RunID,
		"runKey":       job.RunKey,
		"errorMessage": msg,
	}
	if len(job.Params) > 0 {
		var p any
		if err := json.Unmarshal(job.Params, &p); err == nil {
			failureSummary["params"] = p
		}
	}
	failureSummaryJSON, jerr := json.Marshal(failureSummary)
	if jerr == nil {
		var runKeyParam any
		if strings.TrimSpace(job.RunKey) != "" {
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
				summary_json
			)
			VALUES ('market_share', $1::uuid, $2::uuid, 'metrics', 'v1', NULL, $3::jsonb)
			ON CONFLICT (run_kind, run_id, artifact_kind) WHERE deleted_at IS NULL
			DO UPDATE
			SET run_key = EXCLUDED.run_key,
				schema_version = EXCLUDED.schema_version,
				storage_uri = EXCLUDED.storage_uri,
				summary_json = EXCLUDED.summary_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, job.RunID, runKeyParam, failureSummaryJSON)
	}
}
