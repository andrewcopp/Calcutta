package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app/jobqueue"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultSimulationWorkerPollInterval = 5 * time.Second
	defaultSimulationWorkerStaleAfter   = 60 * time.Minute
	simulationWorkerMaxAttempts         = 3
	simulationWorkerConcurrency         = 1
)

// SimulationWorker processes run_simulation jobs from the run_jobs queue.
type SimulationWorker struct {
	pool    *pgxpool.Pool
	claimer *jobqueue.Claimer
}

// NewSimulationWorker creates a new SimulationWorker.
func NewSimulationWorker(pool *pgxpool.Pool) *SimulationWorker {
	return &SimulationWorker{
		pool:    pool,
		claimer: jobqueue.NewClaimer(pool),
	}
}

type runSimulationParams struct {
	Season               int    `json:"season"`
	NSims                int    `json:"nSims"`
	Seed                 int    `json:"seed"`
	StartingStateKey     string `json:"startingStateKey"`
	ProbabilitySourceKey string `json:"probabilitySourceKey"`
}

// Run starts the simulation worker loop.
func (w *SimulationWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultSimulationWorkerPollInterval, defaultSimulationWorkerStaleAfter)
}

// RunWithOptions starts the worker loop with custom settings.
func (w *SimulationWorker) RunWithOptions(ctx context.Context, pollInterval, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		slog.Warn("simulation_worker_disabled", "reason", "database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultSimulationWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultSimulationWorkerStaleAfter
	}

	sem := make(chan struct{}, simulationWorkerConcurrency)
	t := time.NewTicker(pollInterval)
	defer t.Stop()

	kinds := []string{jobqueue.KindRunSimulation}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			select {
			case sem <- struct{}{}:
				job, err := w.claimer.ClaimNext(ctx, kinds, "simulation-worker", simulationWorkerMaxAttempts, staleAfter)
				if err != nil {
					<-sem
					slog.Warn("simulation_worker claim_failed", "error", err)
					continue
				}
				if job == nil {
					<-sem
					continue
				}
				go func(j *jobqueue.Job) {
					defer func() { <-sem }()
					w.processRunSimulation(ctx, j)
				}(job)
			default:
				// At capacity
			}
		}
	}
}

func (w *SimulationWorker) processRunSimulation(ctx context.Context, job *jobqueue.Job) {
	var params runSimulationParams
	if err := json.Unmarshal(job.Params, &params); err != nil {
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, "invalid params: "+err.Error()); failErr != nil {
			slog.Warn("simulation_worker fail_job", "error", failErr)
		}
		return
	}

	if params.Season <= 0 {
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, "missing or invalid season"); failErr != nil {
			slog.Warn("simulation_worker fail_job", "error", failErr)
		}
		return
	}
	if params.NSims <= 0 {
		params.NSims = 10000
	}
	if params.Seed == 0 {
		params.Seed = 42
	}
	if params.StartingStateKey == "" {
		params.StartingStateKey = "current"
	}
	if params.ProbabilitySourceKey == "" {
		params.ProbabilitySourceKey = "lab_pipeline"
	}

	start := time.Now()
	resolver := dbadapters.NewTournamentQueryRepository(w.pool)
	simSvc := simulation.New(w.pool, simulation.WithTournamentResolver(resolver))
	result, err := simSvc.Run(ctx, simulation.RunParams{
		Season:               params.Season,
		NSims:                params.NSims,
		Seed:                 params.Seed,
		Workers:              runtime.GOMAXPROCS(0),
		BatchSize:            1000,
		ProbabilitySourceKey: params.ProbabilitySourceKey,
		StartingStateKey:     params.StartingStateKey,
	})
	if err != nil {
		slog.Warn("simulation_worker run_failed", "season", params.Season, "error", err)
		if failErr := w.claimer.Fail(ctx, job.RunKind, job.RunID, err.Error()); failErr != nil {
			slog.Warn("simulation_worker fail_job", "error", failErr)
		}
		return
	}

	if err := w.claimer.Succeed(ctx, job.RunKind, job.RunID); err != nil {
		slog.Warn("simulation_worker succeed_job", "error", err)
	}

	slog.Info("simulation_worker run_succeeded",
		"season", params.Season,
		"batch_id", result.TournamentSimulationBatchID,
		"n_sims", result.NSims,
		"rows_written", result.RowsWritten,
		"duration_ms", time.Since(start).Milliseconds())
}
