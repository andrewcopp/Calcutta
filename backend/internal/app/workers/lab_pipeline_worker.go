package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/jobqueue"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultLabPipelineWorkerPollInterval = 2 * time.Second
	defaultLabPipelineWorkerStaleAfter   = 30 * time.Minute
	maxConcurrentLabPipelineJobs         = 8
)

// LabPipelineWorkerConfig holds configuration for the lab pipeline worker.
type LabPipelineWorkerConfig struct {
	PythonBin          string
	RunJobsMaxAttempts int
	WorkerID           string
}

// LabPipelineWorker processes lab pipeline jobs (predictions, optimization, evaluation).
type LabPipelineWorker struct {
	pool     *pgxpool.Pool
	progress ProgressWriter
	enqueuer *jobqueue.Enqueuer
	claimer  *jobqueue.Claimer
	sem      chan struct{}
	cfg      LabPipelineWorkerConfig
}

// NewLabPipelineWorker creates a new lab pipeline worker.
func NewLabPipelineWorker(pool *pgxpool.Pool, progress ProgressWriter, cfg LabPipelineWorkerConfig) *LabPipelineWorker {
	if progress == nil {
		progress = NewDBProgressWriter(pool)
	}
	if cfg.PythonBin == "" {
		cfg.PythonBin = "python3"
	}
	if cfg.RunJobsMaxAttempts <= 0 {
		cfg.RunJobsMaxAttempts = 5
	}
	return &LabPipelineWorker{
		pool:     pool,
		progress: progress,
		enqueuer: jobqueue.NewEnqueuer(pool),
		claimer:  jobqueue.NewClaimer(pool),
		sem:      make(chan struct{}, maxConcurrentLabPipelineJobs),
		cfg:      cfg,
	}
}

type labPipelineJob struct {
	RunID     string
	RunKey    string
	RunKind   string
	Params    json.RawMessage
	ClaimedAt time.Time
}

type labPipelineJobParams struct {
	PipelineRunID         string `json:"pipelineRunId"`
	PipelineCalcuttaRunID string `json:"pipelineCalcuttaRunId"`
	InvestmentModelID     string `json:"investmentModelId"`
	CalcuttaID            string `json:"calcuttaId"`
	EntryID               string `json:"entryId"`
	BudgetPoints          int    `json:"budgetPoints"`
	OptimizerKind         string `json:"optimizerKind"`
	NSims                 int    `json:"nSims"`
	Seed                  int    `json:"seed"`
	ExcludedEntryName     string `json:"excludedEntryName"`
}

// Run starts the worker loop.
func (w *LabPipelineWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultLabPipelineWorkerPollInterval, defaultLabPipelineWorkerStaleAfter)
}

// RunWithOptions starts the worker loop with custom poll interval and stale threshold.
func (w *LabPipelineWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		slog.Warn("lab pipeline worker disabled: database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultLabPipelineWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultLabPipelineWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	workerID := w.cfg.WorkerID
	if workerID == "" {
		workerID = "lab-pipeline-worker"
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// Check for pending pipeline runs to kick off
			w.checkAndStartPendingPipelines(ctx)

			// Acquire semaphore before claiming to avoid orphaned jobs
			select {
			case w.sem <- struct{}{}:
				job, ok, err := w.claimNextLabPipelineJob(ctx, workerID, staleAfter)
				if err != nil {
					<-w.sem
					slog.Warn("failed to claim lab pipeline job", "error", err)
					continue
				}
				if !ok {
					<-w.sem
					continue
				}
				go func(j *labPipelineJob) {
					defer func() { <-w.sem }()
					if ok := w.processLabPipelineJob(ctx, workerID, j); !ok {
						slog.Error("lab pipeline job failed", "run_kind", j.RunKind, "run_id", j.RunID)
					}
				}(job)
			default:
				// At capacity, skip this tick
			}
		}
	}
}

func (w *LabPipelineWorker) processLabPipelineJob(ctx context.Context, workerID string, job *labPipelineJob) bool {
	if job == nil {
		return false
	}

	slog.Info("lab_pipeline_worker start", "worker_id", workerID, "run_kind", job.RunKind, "run_id", job.RunID)

	var params labPipelineJobParams
	if err := json.Unmarshal(job.Params, &params); err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("invalid job params: "+err.Error()))
		return false
	}

	var success bool
	switch job.RunKind {
	case "lab_predictions":
		success = w.processPredictionsJob(ctx, workerID, job, params)
	case "lab_optimization":
		success = w.processOptimizationJob(ctx, workerID, job, params)
	case "lab_evaluation":
		success = w.processEvaluationJob(ctx, workerID, job, params)
	default:
		w.failLabPipelineJob(ctx, job, errors.New("unknown run_kind: "+job.RunKind))
		return false
	}

	if success {
		w.checkPipelineCompletion(ctx, params.PipelineRunID)
	}

	return success
}
