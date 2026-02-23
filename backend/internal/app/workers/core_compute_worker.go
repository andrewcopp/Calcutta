package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/jobqueue"
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultCoreComputeWorkerPollInterval = 2 * time.Second
	defaultCoreComputeWorkerStaleAfter   = 10 * time.Minute
	coreComputeWorkerMaxAttempts         = 3
	coreComputeWorkerConcurrency         = 2
)

// CoreComputeWorker processes refresh_predictions jobs from the run_jobs queue.
type CoreComputeWorker struct {
	pool    *pgxpool.Pool
	claimer *jobqueue.Claimer
}

// NewCoreComputeWorker creates a new CoreComputeWorker.
func NewCoreComputeWorker(pool *pgxpool.Pool) *CoreComputeWorker {
	return &CoreComputeWorker{
		pool:    pool,
		claimer: jobqueue.NewClaimer(pool),
	}
}

type refreshPredictionsParams struct {
	TournamentID         string `json:"tournamentId"`
	ProbabilitySourceKey string `json:"probabilitySourceKey"`
}

// Run starts the core compute worker loop.
func (w *CoreComputeWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultCoreComputeWorkerPollInterval, defaultCoreComputeWorkerStaleAfter)
}

// RunWithOptions starts the worker loop with custom settings.
func (w *CoreComputeWorker) RunWithOptions(ctx context.Context, pollInterval, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		slog.Warn("core_compute_worker_disabled", "reason", "database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultCoreComputeWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultCoreComputeWorkerStaleAfter
	}

	sem := make(chan struct{}, coreComputeWorkerConcurrency)
	t := time.NewTicker(pollInterval)
	defer t.Stop()

	kinds := []string{jobqueue.KindRefreshPredictions}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			select {
			case sem <- struct{}{}:
				job, err := w.claimer.ClaimNext(ctx, kinds, "core-compute-worker", coreComputeWorkerMaxAttempts, staleAfter)
				if err != nil {
					<-sem
					slog.Warn("core_compute_worker claim_failed", "error", err)
					continue
				}
				if job == nil {
					<-sem
					continue
				}
				go func(j *jobqueue.Job) {
					defer func() { <-sem }()
					w.processRefreshPredictions(ctx, j)
				}(job)
			default:
				// At capacity
			}
		}
	}
}

func (w *CoreComputeWorker) processRefreshPredictions(ctx context.Context, job *jobqueue.Job) {
	var params refreshPredictionsParams
	if err := json.Unmarshal(job.Params, &params); err != nil {
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, "invalid params: "+err.Error()); failErr != nil {
			slog.Warn("core_compute_worker fail_job", "error", failErr)
		}
		return
	}

	if params.TournamentID == "" {
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, "missing tournamentId"); failErr != nil {
			slog.Warn("core_compute_worker fail_job", "error", failErr)
		}
		return
	}

	sourceKey := params.ProbabilitySourceKey
	if sourceKey == "" {
		sourceKey = "kenpom"
	}

	start := time.Now()
	predSvc := prediction.New(w.pool)
	results, err := predSvc.RunAllCheckpoints(ctx, prediction.RunParams{
		TournamentID:         params.TournamentID,
		ProbabilitySourceKey: sourceKey,
	})
	if err != nil {
		slog.Warn("core_compute_worker prediction_failed", "tournament_id", params.TournamentID, "error", err)
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, err.Error()); failErr != nil {
			slog.Warn("core_compute_worker fail_job", "error", failErr)
		}
		return
	}

	if err := w.claimer.Succeed(ctx, job.RunKind, job.RunID); err != nil {
		slog.Warn("core_compute_worker succeed_job", "error", err)
	}

	for _, result := range results {
		slog.Info("core_compute_worker prediction_succeeded",
			"tournament_id", params.TournamentID,
			"batch_id", result.BatchID,
			"team_count", result.TeamCount,
			"duration_ms", result.Duration.Milliseconds())
	}
	slog.Info("core_compute_worker all_checkpoints_done",
		"tournament_id", params.TournamentID,
		"checkpoints", len(results),
		"total_duration_ms", time.Since(start).Milliseconds())
}
