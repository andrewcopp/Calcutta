package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	appcalcuttaevaluations "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta_evaluations"
	"github.com/andrewcopp/Calcutta/backend/internal/app/recommended_entry_bids"
)

const (
	defaultLabPipelineWorkerPollInterval = 2 * time.Second
	defaultLabPipelineWorkerStaleAfter   = 30 * time.Minute
	maxConcurrentLabPipelineJobs         = 8
)

// LabPipelineWorker processes lab pipeline jobs (predictions, optimization, evaluation).
type LabPipelineWorker struct {
	pool     *pgxpool.Pool
	progress ProgressWriter
	sem      chan struct{}
}

// NewLabPipelineWorker creates a new lab pipeline worker.
func NewLabPipelineWorker(pool *pgxpool.Pool, progress ProgressWriter) *LabPipelineWorker {
	if progress == nil {
		progress = NewDBProgressWriter(pool)
	}
	return &LabPipelineWorker{
		pool:     pool,
		progress: progress,
		sem:      make(chan struct{}, maxConcurrentLabPipelineJobs),
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
	PipelineRunID         string `json:"pipeline_run_id"`
	PipelineCalcuttaRunID string `json:"pipeline_calcutta_run_id"`
	InvestmentModelID     string `json:"investment_model_id"`
	CalcuttaID            string `json:"calcutta_id"`
	EntryID               string `json:"entry_id"`
	BudgetPoints          int    `json:"budget_points"`
	OptimizerKind         string `json:"optimizer_kind"`
	NSims                 int    `json:"n_sims"`
	Seed                  int    `json:"seed"`
	ExcludedEntryName     string `json:"excluded_entry_name"`
}

// Run starts the worker loop.
func (w *LabPipelineWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultLabPipelineWorkerPollInterval, defaultLabPipelineWorkerStaleAfter)
}

// RunWithOptions starts the worker loop with custom poll interval and stale threshold.
func (w *LabPipelineWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		log.Printf("lab pipeline worker disabled: database pool not available")
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

	workerID := os.Getenv("HOSTNAME")
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
					log.Printf("Error claiming next lab pipeline job: %v", err)
					continue
				}
				if !ok {
					<-w.sem
					continue
				}
				go func(j *labPipelineJob) {
					defer func() { <-w.sem }()
					_ = w.processLabPipelineJob(ctx, workerID, j)
				}(job)
			default:
				// At capacity, skip this tick
			}
		}
	}
}

func (w *LabPipelineWorker) checkAndStartPendingPipelines(ctx context.Context) {
	// Find pending pipeline runs and enqueue their first jobs
	rows, err := w.pool.Query(ctx, `
		SELECT pr.id::text, pr.investment_model_id::text, pr.budget_points, pr.optimizer_kind,
		       pr.n_sims, pr.seed, pr.excluded_entry_name
		FROM lab.pipeline_runs pr
		WHERE pr.status = 'pending'
		ORDER BY pr.created_at ASC
		LIMIT 5
	`)
	if err != nil {
		log.Printf("lab_pipeline_worker check_pending error=%v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pipelineRunID, modelID string
		var budgetPoints, nSims, seed int
		var optimizerKind string
		var excludedEntryName *string
		if err := rows.Scan(&pipelineRunID, &modelID, &budgetPoints, &optimizerKind, &nSims, &seed, &excludedEntryName); err != nil {
			log.Printf("lab_pipeline_worker scan error=%v", err)
			continue
		}

		// Update pipeline to running
		_, err := w.pool.Exec(ctx, `
			UPDATE lab.pipeline_runs
			SET status = 'running', started_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND status = 'pending'
		`, pipelineRunID)
		if err != nil {
			log.Printf("lab_pipeline_worker update_running error=%v", err)
			continue
		}

		// Get all calcutta runs for this pipeline and enqueue prediction jobs
		calcuttaRows, err := w.pool.Query(ctx, `
			SELECT id::text, calcutta_id::text
			FROM lab.pipeline_calcutta_runs
			WHERE pipeline_run_id = $1::uuid AND status = 'pending'
		`, pipelineRunID)
		if err != nil {
			log.Printf("lab_pipeline_worker get_calcuttas error=%v", err)
			continue
		}

		for calcuttaRows.Next() {
			var pcrID, calcuttaID string
			if err := calcuttaRows.Scan(&pcrID, &calcuttaID); err != nil {
				continue
			}

			// Create job params
			params := labPipelineJobParams{
				PipelineRunID:         pipelineRunID,
				PipelineCalcuttaRunID: pcrID,
				InvestmentModelID:     modelID,
				CalcuttaID:            calcuttaID,
				BudgetPoints:          budgetPoints,
				OptimizerKind:         optimizerKind,
				NSims:                 nSims,
				Seed:                  seed,
			}
			if excludedEntryName != nil {
				params.ExcludedEntryName = *excludedEntryName
			}
			paramsJSON, err := json.Marshal(params)
			if err != nil {
				log.Printf("lab_pipeline_worker marshal_params error=%v", err)
				continue
			}

			// Enqueue prediction job
			var jobID string
			err = w.pool.QueryRow(ctx, `
				INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
				VALUES ('lab_predictions', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
				RETURNING run_id::text
			`, pcrID, paramsJSON).Scan(&jobID)
			if err != nil {
				log.Printf("lab_pipeline_worker enqueue_job error=%v", err)
				continue
			}

			// Update calcutta run with job ID
			_, _ = w.pool.Exec(ctx, `
				UPDATE lab.pipeline_calcutta_runs
				SET predictions_job_id = $2::uuid, status = 'running', started_at = NOW(), updated_at = NOW()
				WHERE id = $1::uuid
			`, pcrID, jobID)

			log.Printf("lab_pipeline_worker enqueued_predictions pipeline_run=%s calcutta_run=%s job_id=%s", pipelineRunID, pcrID, jobID)
		}
		calcuttaRows.Close()
	}
}

func (w *LabPipelineWorker) claimNextLabPipelineJob(ctx context.Context, workerID string, staleAfter time.Duration) (*labPipelineJob, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	maxAttempts := resolveRunJobsMaxAttempts(5)
	baseStaleSeconds := staleAfter.Seconds()
	if baseStaleSeconds <= 0 {
		baseStaleSeconds = defaultLabPipelineWorkerStaleAfter.Seconds()
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Fail stale jobs that exceeded max attempts
	if _, err := tx.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed',
			finished_at = NOW(),
			error_message = COALESCE(error_message, 'max_attempts_exceeded'),
			updated_at = NOW()
		WHERE run_kind IN ('lab_predictions', 'lab_optimization', 'lab_evaluation')
			AND status = 'running'
			AND claimed_at IS NOT NULL
			AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
			AND attempt >= $3
	`, pgtype.Timestamptz{Time: now, Valid: true}, baseStaleSeconds, maxAttempts); err != nil {
		return nil, false, err
	}

	// Claim next job
	q := `
		WITH candidate AS (
			SELECT id
			FROM derived.run_jobs
			WHERE run_kind IN ('lab_predictions', 'lab_optimization', 'lab_evaluation')
				AND attempt < $4
				AND (
					status = 'queued'
					OR (
						status = 'running'
						AND claimed_at IS NOT NULL
						AND claimed_at < ($1::timestamptz - make_interval(secs => ($2 * POWER(2, GREATEST(attempt - 1, 0)))))
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
		RETURNING j.run_id::text, j.run_key::text, j.run_kind, j.params_json::text
	`

	job := &labPipelineJob{}
	var paramsStr string
	if err := tx.QueryRow(ctx, q,
		pgtype.Timestamptz{Time: now, Valid: true},
		baseStaleSeconds,
		workerID,
		maxAttempts,
	).Scan(&job.RunID, &job.RunKey, &job.RunKind, &paramsStr); err != nil {
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

func (w *LabPipelineWorker) processLabPipelineJob(ctx context.Context, workerID string, job *labPipelineJob) bool {
	if job == nil {
		return false
	}

	log.Printf("lab_pipeline_worker start worker_id=%s run_kind=%s run_id=%s", workerID, job.RunKind, job.RunID)

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

func (w *LabPipelineWorker) processPredictionsJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.1, "predictions", "Generating market predictions")

	// Get model kind to determine which script to use
	var modelKind string
	err := w.pool.QueryRow(ctx, `
		SELECT kind FROM lab.investment_models WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.InvestmentModelID).Scan(&modelKind)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to get model kind: %w", err))
		return false
	}

	// Run Python script
	pythonBin := strings.TrimSpace(os.Getenv("PYTHON_BIN"))
	if pythonBin == "" {
		pythonBin = "python3"
	}

	// Choose script based on model kind
	var scriptName string
	switch modelKind {
	case "oracle":
		scriptName = "data-science/scripts/generate_oracle_predictions.py"
	case "naive_ev":
		scriptName = "data-science/scripts/generate_naive_ev_predictions.py"
	default:
		scriptName = "data-science/scripts/generate_lab_predictions.py"
	}

	scriptPath := w.resolvePythonScript(scriptName)
	if scriptPath == "" {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("predictions script not found: %s", scriptName))
		return false
	}

	cmd := exec.CommandContext(ctx, pythonBin, scriptPath,
		"--model-id", params.InvestmentModelID,
		"--calcutta-id", params.CalcuttaID,
		"--json-output",
	)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	dur := time.Since(start)

	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		w.failLabPipelineJob(ctx, job, errors.New(errMsg))
		log.Printf("lab_pipeline_worker predictions_fail worker_id=%s run_id=%s dur_ms=%d err=%s", workerID, job.RunID, dur.Milliseconds(), errMsg)
		return false
	}

	// Parse output to get entry ID
	var result struct {
		OK      bool   `json:"ok"`
		EntryID string `json:"entry_id"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || !result.OK {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = "predictions script returned error"
		}
		w.failLabPipelineJob(ctx, job, errors.New(errMsg))
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "predictions", "Predictions complete")

	// Mark job succeeded
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run and enqueue next stage
	_, _ = w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET entry_id = $2::uuid, stage = 'optimization', progress = 0.33, updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID, result.EntryID)

	// Enqueue optimization job
	nextParams := params
	nextParams.EntryID = result.EntryID
	nextParamsJSON, _ := json.Marshal(nextParams)

	var nextJobID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
		VALUES ('lab_optimization', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
		RETURNING run_id::text
	`, params.PipelineCalcuttaRunID, nextParamsJSON).Scan(&nextJobID)
	if err != nil {
		log.Printf("lab_pipeline_worker enqueue_optimization error=%v", err)
	} else {
		_, _ = w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET optimization_job_id = $2::uuid, updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, nextJobID)
	}

	log.Printf("lab_pipeline_worker predictions_success worker_id=%s run_id=%s entry_id=%s dur_ms=%d", workerID, job.RunID, result.EntryID, dur.Milliseconds())
	return true
}

func (w *LabPipelineWorker) processOptimizationJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.4, "optimization", "Optimizing bids with DP allocator")

	start := time.Now()

	// Fetch predictions from database
	var predictionsJSON []byte
	err := w.pool.QueryRow(ctx, `
		SELECT predictions_json FROM lab.entries
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.EntryID).Scan(&predictionsJSON)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to fetch predictions: %w", err))
		return false
	}
	if len(predictionsJSON) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("entry has no predictions - cannot optimize"))
		return false
	}

	// Parse predictions
	type prediction struct {
		TeamID               string  `json:"team_id"`
		PredictedMarketShare float64 `json:"predicted_market_share"`
		ExpectedPoints       float64 `json:"expected_points"`
	}
	var predictions []prediction
	if err := json.Unmarshal(predictionsJSON, &predictions); err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to parse predictions: %w", err))
		return false
	}
	if len(predictions) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("predictions array is empty - cannot optimize"))
		return false
	}

	// Get calcutta rules for constraints
	var minTeams, maxTeams, maxPerTeam int32
	err = w.pool.QueryRow(ctx, `
		SELECT min_teams, max_teams, max_bid
		FROM core.calcuttas
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.CalcuttaID).Scan(&minTeams, &maxTeams, &maxPerTeam)
	if err != nil {
		// Use sensible defaults if calcutta not found
		minTeams, maxTeams, maxPerTeam = 3, 10, 50
		log.Printf("lab_pipeline_worker optimization_warn using default constraints: %v", err)
	}

	// Get total pool budget (number of entries × budget per entry)
	// This is needed because predicted_market_share is a fraction of the TOTAL pool
	var totalPoolBudget int
	err = w.pool.QueryRow(ctx, `
		SELECT c.budget_points * COUNT(e.id)::int
		FROM core.calcuttas c
		LEFT JOIN core.entries e ON e.calcutta_id = c.id AND e.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_points
	`, params.CalcuttaID).Scan(&totalPoolBudget)
	if err != nil || totalPoolBudget <= 0 {
		// Fallback: estimate from number of entries * budget_points
		totalPoolBudget = 4200 // reasonable default for ~42 entries * 100 budget
		log.Printf("lab_pipeline_worker optimization_warn using default total pool budget: %v", err)
	}

	// Build teams for allocator
	budgetPoints := params.BudgetPoints
	if budgetPoints <= 0 {
		budgetPoints = 100
	}

	teams := make([]recommended_entry_bids.Team, len(predictions))
	for i, pred := range predictions {
		teams[i] = recommended_entry_bids.Team{
			ID:             pred.TeamID,
			ExpectedPoints: pred.ExpectedPoints,
			MarketPoints:   pred.PredictedMarketShare * float64(totalPoolBudget),
		}
	}

	// Run the Go DP allocator
	allocParams := recommended_entry_bids.AllocationParams{
		BudgetPoints: budgetPoints,
		MinTeams:     int(minTeams),
		MaxTeams:     int(maxTeams),
		MinBidPoints: 1,
		MaxBidPoints: int(maxPerTeam),
	}
	result, err := recommended_entry_bids.AllocateBids(teams, allocParams)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("allocator failed: %w", err))
		return false
	}

	// FAIL FAST: Validate the allocation
	totalBid := 0
	for _, bid := range result.Bids {
		totalBid += bid
	}
	numTeams := len(result.Bids)

	// Strict validation - no silent fallbacks
	if totalBid > budgetPoints {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated budget constraint: total=%d > budget=%d", totalBid, budgetPoints))
		return false
	}
	if numTeams > 0 && numTeams < int(minTeams) {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated min_teams constraint: count=%d < min=%d", numTeams, minTeams))
		return false
	}
	if numTeams > int(maxTeams) {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated max_teams constraint: count=%d > max=%d", numTeams, maxTeams))
		return false
	}
	for teamID, bid := range result.Bids {
		if bid > int(maxPerTeam) {
			w.failLabPipelineJob(ctx, job, fmt.Errorf("CRITICAL: allocator violated max_per_team constraint: team=%s bid=%d > max=%d", teamID, bid, maxPerTeam))
			return false
		}
	}

	// Build bids JSON with expected ROI
	type bidRow struct {
		TeamID      string  `json:"team_id"`
		BidPoints   int     `json:"bid_points"`
		ExpectedROI float64 `json:"expected_roi"`
	}
	bids := make([]bidRow, 0, len(result.Bids))
	for _, pred := range predictions {
		bid, ok := result.Bids[pred.TeamID]
		if !ok || bid == 0 {
			continue
		}
		marketCost := pred.PredictedMarketShare * float64(budgetPoints)
		expectedROI := 0.0
		if (marketCost + float64(bid)) > 0 {
			expectedROI = pred.ExpectedPoints / (marketCost + float64(bid))
		}
		bids = append(bids, bidRow{
			TeamID:      pred.TeamID,
			BidPoints:   bid,
			ExpectedROI: expectedROI,
		})
	}

	bidsJSON, err := json.Marshal(bids)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to marshal bids: %w", err))
		return false
	}

	// Save bids to database
	optimizerParams := map[string]interface{}{
		"budget_points": budgetPoints,
		"min_teams":     minTeams,
		"max_teams":     maxTeams,
		"max_per_team":  maxPerTeam,
		"min_bid":       1,
	}
	optimizerParamsJSON, _ := json.Marshal(optimizerParams)

	_, err = w.pool.Exec(ctx, `
		UPDATE lab.entries
		SET bids_json = $2::jsonb,
			optimizer_kind = 'dp',
			optimizer_params_json = $3::jsonb,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, params.EntryID, bidsJSON, optimizerParamsJSON)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to save bids: %w", err))
		return false
	}

	dur := time.Since(start)

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "optimization", "Optimization complete")
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run and enqueue evaluation
	_, _ = w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET stage = 'evaluation', progress = 0.66, updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID)

	// Enqueue evaluation job
	nextParamsJSON, _ := json.Marshal(params)
	var nextJobID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO derived.run_jobs (run_kind, run_id, run_key, params_json, status)
		VALUES ('lab_evaluation', uuid_generate_v4(), $1::uuid, $2::jsonb, 'queued')
		RETURNING run_id::text
	`, params.PipelineCalcuttaRunID, nextParamsJSON).Scan(&nextJobID)
	if err != nil {
		log.Printf("lab_pipeline_worker enqueue_evaluation error=%v", err)
	} else {
		_, _ = w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET evaluation_job_id = $2::uuid, updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, nextJobID)
	}

	log.Printf("lab_pipeline_worker optimization_success worker_id=%s run_id=%s teams=%d total_bid=%d dur_ms=%d", workerID, job.RunID, numTeams, totalBid, dur.Milliseconds())
	return true
}

func (w *LabPipelineWorker) processEvaluationJob(ctx context.Context, workerID string, job *labPipelineJob, params labPipelineJobParams) bool {
	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.7, "evaluation", "Running simulation")

	start := time.Now()

	// Get entry bids and calcutta_id
	var bidsJSON []byte
	var calcuttaID string
	err := w.pool.QueryRow(ctx, `
		SELECT bids_json, calcutta_id::text FROM lab.entries WHERE id = $1::uuid AND deleted_at IS NULL
	`, params.EntryID).Scan(&bidsJSON, &calcuttaID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to get entry bids: "+err.Error()))
		return false
	}

	// Parse bids into map for evaluation
	type bidEntry struct {
		TeamID    string `json:"team_id"`
		BidPoints int    `json:"bid_points"`
	}
	var bids []bidEntry
	if err := json.Unmarshal(bidsJSON, &bids); err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to parse bids: "+err.Error()))
		return false
	}

	labEntryBids := make(map[string]int)
	for _, b := range bids {
		if b.BidPoints > 0 {
			labEntryBids[b.TeamID] = b.BidPoints
		}
	}

	if len(labEntryBids) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("entry has no bids to evaluate"))
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.75, "evaluation", "Running "+fmt.Sprintf("%d", params.NSims)+" simulations")

	// Run evaluation using calcutta_evaluations service
	evalService := appcalcuttaevaluations.New(w.pool)
	result, err := evalService.EvaluateLabEntry(ctx, calcuttaID, labEntryBids, params.ExcludedEntryName)
	if err != nil {
		w.failLabPipelineJob(ctx, job, fmt.Errorf("evaluation failed: %w", err))
		return false
	}

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 0.95, "evaluation", "Saving results")

	// Extract "Our Strategy" rank from results
	var ourRank int
	for _, entry := range result.AllEntryResults {
		if entry.EntryName == "Our Strategy" {
			ourRank = entry.Rank
			break
		}
	}

	// Create or update lab.evaluations row with results
	var evaluationID string
	err = w.pool.QueryRow(ctx, `
		INSERT INTO lab.evaluations (entry_id, n_sims, seed, mean_normalized_payout, median_normalized_payout, p_top1, p_in_money, our_rank)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (entry_id, n_sims, seed) WHERE deleted_at IS NULL
		DO UPDATE SET
			mean_normalized_payout = EXCLUDED.mean_normalized_payout,
			median_normalized_payout = EXCLUDED.median_normalized_payout,
			p_top1 = EXCLUDED.p_top1,
			p_in_money = EXCLUDED.p_in_money,
			our_rank = EXCLUDED.our_rank,
			updated_at = NOW()
		RETURNING id::text
	`, params.EntryID, result.NSims, params.Seed, result.MeanNormalizedPayout, result.MedianNormalizedPayout, result.PTop1, result.PInMoney, ourRank).Scan(&evaluationID)
	if err != nil {
		w.failLabPipelineJob(ctx, job, errors.New("failed to save evaluation: "+err.Error()))
		return false
	}

	// Save per-entry results
	if len(result.AllEntryResults) == 0 {
		w.failLabPipelineJob(ctx, job, errors.New("evaluation produced no entry results"))
		return false
	}
	{
		// Delete existing results for this evaluation (in case of re-run)
		_, _ = w.pool.Exec(ctx, `
			DELETE FROM lab.evaluation_entry_results WHERE evaluation_id = $1::uuid
		`, evaluationID)

		// Insert all entry results
		for _, entry := range result.AllEntryResults {
			_, err := w.pool.Exec(ctx, `
				INSERT INTO lab.evaluation_entry_results (evaluation_id, entry_name, mean_normalized_payout, p_top1, p_in_money, rank)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)
			`, evaluationID, entry.EntryName, entry.MeanPayout, entry.PTop1, entry.PInMoney, entry.Rank)
			if err != nil {
				w.failLabPipelineJob(ctx, job, fmt.Errorf("failed to save entry result for %s: %w", entry.EntryName, err))
				return false
			}
		}
	}

	dur := time.Since(start)

	w.updateProgress(ctx, job.RunKind, job.RunID, params.PipelineCalcuttaRunID, 1.0, "evaluation", "Evaluation complete")
	w.succeedLabPipelineJob(ctx, job)

	// Update calcutta run as completed
	_, _ = w.pool.Exec(ctx, `
		UPDATE lab.pipeline_calcutta_runs
		SET stage = 'completed', status = 'succeeded', progress = 1.0,
		    evaluation_id = $2::uuid, finished_at = NOW(), updated_at = NOW()
		WHERE id = $1::uuid
	`, params.PipelineCalcuttaRunID, evaluationID)

	// Update entry state to complete
	_, _ = w.pool.Exec(ctx, `
		UPDATE lab.entries SET state = 'complete', updated_at = NOW() WHERE id = $1::uuid
	`, params.EntryID)

	log.Printf("lab_pipeline_worker evaluation_success worker_id=%s run_id=%s evaluation_id=%s n_sims=%d mean_payout=%.4f p_top1=%.4f dur_ms=%d",
		workerID, job.RunID, evaluationID, result.NSims, result.MeanNormalizedPayout, result.PTop1, dur.Milliseconds())
	return true
}

func (w *LabPipelineWorker) checkPipelineCompletion(ctx context.Context, pipelineRunID string) {
	// Check if all calcutta runs are complete
	var pending, running, failed, succeeded int
	err := w.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'running'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'succeeded')
		FROM lab.pipeline_calcutta_runs
		WHERE pipeline_run_id = $1::uuid
	`, pipelineRunID).Scan(&pending, &running, &failed, &succeeded)
	if err != nil {
		log.Printf("lab_pipeline_worker check_completion error=%v", err)
		return
	}

	if pending > 0 || running > 0 {
		return // Still processing
	}

	// All done - update pipeline status
	status := "succeeded"
	if failed > 0 && succeeded == 0 {
		status = "failed"
	} else if failed > 0 {
		status = "succeeded" // Partial success
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE lab.pipeline_runs
		SET status = $2, finished_at = NOW(), updated_at = NOW()
		WHERE id = $1::uuid AND status = 'running'
	`, pipelineRunID, status)

	log.Printf("lab_pipeline_worker pipeline_complete pipeline_run=%s status=%s succeeded=%d failed=%d", pipelineRunID, status, succeeded, failed)
}

func (w *LabPipelineWorker) updateProgress(ctx context.Context, runKind, runID, pcrID string, percent float64, phase, message string) {
	if w.progress != nil {
		w.progress.Update(ctx, runKind, runID, percent, phase, message)
	}

	// Also update pipeline_calcutta_runs progress
	if pcrID != "" {
		_, _ = w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET progress = $2, progress_message = $3, updated_at = NOW()
			WHERE id = $1::uuid
		`, pcrID, percent, message)
	}
}

func (w *LabPipelineWorker) succeedLabPipelineJob(ctx context.Context, job *labPipelineJob) {
	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'succeeded', finished_at = NOW(), error_message = NULL, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, job.RunKind, job.RunID)
}

func (w *LabPipelineWorker) failLabPipelineJob(ctx context.Context, job *labPipelineJob, err error) {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}

	_, _ = w.pool.Exec(ctx, `
		UPDATE derived.run_jobs
		SET status = 'failed', finished_at = NOW(), error_message = $3, updated_at = NOW()
		WHERE run_kind = $1 AND run_id = $2::uuid
	`, job.RunKind, job.RunID, msg)

	// Also update pipeline_calcutta_runs
	var params labPipelineJobParams
	if err := json.Unmarshal(job.Params, &params); err == nil && params.PipelineCalcuttaRunID != "" {
		_, _ = w.pool.Exec(ctx, `
			UPDATE lab.pipeline_calcutta_runs
			SET status = 'failed', error_message = $2, finished_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid
		`, params.PipelineCalcuttaRunID, msg)
	}

	if w.progress != nil {
		w.progress.Update(ctx, job.RunKind, job.RunID, 1.0, "failed", msg)
	}
}

func (w *LabPipelineWorker) resolvePythonScript(relativePath string) string {
	candidates := []string{
		relativePath,
		"../" + relativePath,
		"../../" + relativePath,
	}

	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}
