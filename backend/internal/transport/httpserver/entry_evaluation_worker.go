package httpserver

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/features/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultEntryEvalWorkerPollInterval = 2 * time.Second
	defaultEntryEvalWorkerStaleAfter   = 30 * time.Minute
)

type entryEvaluationRequestRow struct {
	ID               string
	CalcuttaID       string
	EntryCandidateID string
	ExcludedEntry    *string
	StartingStateKey string
	NSims            int
	Seed             int
	ExperimentKey    string
	RequestSource    string
}

func (s *Server) RunEntryEvaluationWorker(ctx context.Context) {
	s.RunEntryEvaluationWorkerWithOptions(ctx, defaultEntryEvalWorkerPollInterval, defaultEntryEvalWorkerStaleAfter)
}

func (s *Server) RunEntryEvaluationWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("Entry evaluation worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultEntryEvalWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultEntryEvalWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := os.Getenv("HOSTNAME")
	if workerID == "" {
		workerID = "entry-eval-worker"
	}

	claimedCount := int64(0)
	succeededCount := int64(0)
	failedCount := int64(0)
	lastStatsAt := time.Now().UTC()
	lastStatsClaimed := int64(0)
	lastStatsSucceeded := int64(0)
	lastStatsFailed := int64(0)

	var totalJobDuration time.Duration
	var totalSimDuration time.Duration
	var totalEvalDuration time.Duration

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			req, ok, err := s.claimNextEntryEvaluationRequest(ctx, workerID, staleAfter)
			if err != nil {
				log.Printf("Error claiming next entry evaluation request: %v", err)
				continue
			}
			if !ok {
				if time.Since(lastStatsAt) >= 60*time.Second {
					deltaClaimed := claimedCount - lastStatsClaimed
					deltaSucceeded := succeededCount - lastStatsSucceeded
					deltaFailed := failedCount - lastStatsFailed

					avgJobMs := int64(0)
					avgSimMs := int64(0)
					avgEvalMs := int64(0)
					if claimedCount > 0 {
						avgJobMs = int64(totalJobDuration.Milliseconds()) / claimedCount
						avgSimMs = int64(totalSimDuration.Milliseconds()) / claimedCount
						avgEvalMs = int64(totalEvalDuration.Milliseconds()) / claimedCount
					}

					log.Printf("entry_eval_worker stats worker_id=%s claimed_total=%d succeeded_total=%d failed_total=%d claimed_1m=%d succeeded_1m=%d failed_1m=%d avg_job_ms=%d avg_sim_ms=%d avg_eval_ms=%d",
						workerID,
						claimedCount,
						succeededCount,
						failedCount,
						deltaClaimed,
						deltaSucceeded,
						deltaFailed,
						avgJobMs,
						avgSimMs,
						avgEvalMs,
					)

					lastStatsAt = time.Now().UTC()
					lastStatsClaimed = claimedCount
					lastStatsSucceeded = succeededCount
					lastStatsFailed = failedCount
				}
				continue
			}

			claimedCount++
			jobStart := time.Now()
			simDur, evalDur, ok := s.processEntryEvaluationRequest(ctx, workerID, req)
			totalJobDuration += time.Since(jobStart)
			totalSimDuration += simDur
			totalEvalDuration += evalDur
			if ok {
				succeededCount++
			} else {
				failedCount++
			}
		}
	}
}

func (s *Server) claimNextEntryEvaluationRequest(ctx context.Context, workerID string, staleAfter time.Duration) (*entryEvaluationRequestRow, bool, error) {
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

	row := &entryEvaluationRequestRow{}
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.entry_evaluation_requests
			WHERE deleted_at IS NULL
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
		UPDATE derived.entry_evaluation_requests r
		SET status = 'running',
			claimed_at = $1,
			claimed_by = $3,
			error_message = NULL
		FROM candidate
		WHERE r.id = candidate.id
		RETURNING
			r.id,
			r.calcutta_id,
			r.entry_candidate_id,
			r.excluded_entry_name,
			r.starting_state_key,
			r.n_sims,
			r.seed,
			r.experiment_key,
			r.request_source
	`

	var excluded *string
	if err := tx.QueryRow(ctx, q, pgtype.Timestamptz{Time: now, Valid: true}, pgtype.Timestamptz{Time: staleBefore, Valid: true}, workerID).Scan(
		&row.ID,
		&row.CalcuttaID,
		&row.EntryCandidateID,
		&excluded,
		&row.StartingStateKey,
		&row.NSims,
		&row.Seed,
		&row.ExperimentKey,
		&row.RequestSource,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	row.ExcludedEntry = excluded

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	committed = true

	return row, true, nil
}

func (s *Server) processEntryEvaluationRequest(ctx context.Context, workerID string, req *entryEvaluationRequestRow) (time.Duration, time.Duration, bool) {
	if req == nil {
		return 0, 0, false
	}

	runID := uuid.NewString()

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	log.Printf("entry_eval_worker start worker_id=%s request_id=%s run_id=%s calcutta_id=%s entry_candidate_id=%s experiment_key=%s request_source=%s starting_state_key=%s n_sims=%d seed=%d excluded_entry_name=%q",
		workerID,
		req.ID,
		runID,
		req.CalcuttaID,
		req.EntryCandidateID,
		req.ExperimentKey,
		req.RequestSource,
		req.StartingStateKey,
		req.NSims,
		req.Seed,
		excluded,
	)

	year, err := s.resolveSeasonYearByCalcuttaID(ctx, req.CalcuttaID)
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_id=%s err=%v", workerID, req.ID, runID, err)
		return 0, 0, false
	}

	simSvc := appsimulatetournaments.New(s.pool)
	simStart := time.Now()
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                req.NSims,
		Seed:                 req.Seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "entry_eval_worker",
		StartingStateKey:     req.StartingStateKey,
	})
	simDur := time.Since(simStart)
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_id=%s phase=simulate err=%v", workerID, req.ID, runID, err)
		return simDur, 0, false
	}

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalStart := time.Now()
	evalRunID, err := evalSvc.CalculateSimulatedCalcuttaForEntryCandidate(
		ctx,
		req.CalcuttaID,
		runID,
		excluded,
		&simRes.TournamentSimulationBatchID,
		req.EntryCandidateID,
	)
	evalDur := time.Since(evalStart)
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		log.Printf("entry_eval_worker fail worker_id=%s request_id=%s run_id=%s phase=evaluate err=%v", workerID, req.ID, runID, err)
		return simDur, evalDur, false
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE derived.entry_evaluation_requests
		SET status = 'succeeded',
			evaluation_run_id = $2::uuid,
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, req.ID, evalRunID)
	if err != nil {
		log.Printf("Error updating entry evaluation request %s to succeeded: %v", req.ID, err)
		return simDur, evalDur, false
	}

	log.Printf("entry_eval_worker success worker_id=%s request_id=%s run_id=%s evaluation_run_id=%s sim_ms=%d eval_ms=%d",
		workerID,
		req.ID,
		runID,
		evalRunID,
		simDur.Milliseconds(),
		evalDur.Milliseconds(),
	)
	return simDur, evalDur, true
}

func (s *Server) failEntryEvaluationRequest(ctx context.Context, requestID string, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	_, e := s.pool.Exec(ctx, `
		UPDATE derived.entry_evaluation_requests
		SET status = 'failed',
			error_message = $2,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, requestID, msg)
	if e != nil {
		log.Printf("Error marking entry evaluation request %s failed: %v (original error: %v)", requestID, e, err)
	}
}

func (s *Server) resolveSeasonYearByCalcuttaID(ctx context.Context, calcuttaID string) (int, error) {
	var year int
	q := `
		SELECT seas.year
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id AND seas.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`
	if err := s.pool.QueryRow(ctx, q, calcuttaID).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
}
