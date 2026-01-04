package httpserver

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	appsimulatetournaments "github.com/andrewcopp/Calcutta/backend/internal/features/simulate_tournaments"
	appsimulatedcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/simulated_calcutta"
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
				continue
			}
			s.processEntryEvaluationRequest(ctx, workerID, req)
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
			r.seed
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

func (s *Server) processEntryEvaluationRequest(ctx context.Context, workerID string, req *entryEvaluationRequestRow) {
	if req == nil {
		return
	}

	excluded := ""
	if req.ExcludedEntry != nil {
		excluded = *req.ExcludedEntry
	}

	year, err := s.resolveSeasonYearByCalcuttaID(ctx, req.CalcuttaID)
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		return
	}

	simSvc := appsimulatetournaments.New(s.pool)
	simRes, err := simSvc.Run(ctx, appsimulatetournaments.RunParams{
		Season:               year,
		NSims:                req.NSims,
		Seed:                 req.Seed,
		Workers:              0,
		BatchSize:            500,
		ProbabilitySourceKey: "entry_eval_worker",
		StartingStateKey:     req.StartingStateKey,
	})
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		return
	}

	evalSvc := appsimulatedcalcutta.New(s.pool)
	evalRunID, err := evalSvc.CalculateSimulatedCalcuttaForEntryCandidate(
		ctx,
		simRes.LabTournamentID,
		req.ID,
		excluded,
		&simRes.TournamentSimulationBatchID,
		req.EntryCandidateID,
	)
	if err != nil {
		s.failEntryEvaluationRequest(ctx, req.ID, err)
		return
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
		return
	}
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
