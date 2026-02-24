package lab

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// ServiceConfig holds configuration for the lab service.
type ServiceConfig struct {
	DefaultNSims      int
	ExcludedEntryName string
}

// Service provides lab-related business logic.
type Service struct {
	repo         ports.LabRepository
	pipelineRepo ports.LabPipelineRepository
	cfg          ServiceConfig
}

// New creates a new lab service.
func New(repo ports.LabPipelineRepository, cfg ServiceConfig) *Service {
	return &Service{repo: repo, pipelineRepo: repo, cfg: cfg}
}

// clampPagination enforces default and maximum bounds on pagination parameters.
func clampPagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// ListInvestmentModels returns investment models matching the filter.
func (s *Service) ListInvestmentModels(ctx context.Context, filter models.LabListModelsFilter, page models.LabPagination) ([]models.InvestmentModel, error) {
	page.Limit, page.Offset = clampPagination(page.Limit, page.Offset)
	return s.repo.ListInvestmentModels(ctx, filter, page)
}

// GetInvestmentModel returns a single investment model by ID.
func (s *Service) GetInvestmentModel(ctx context.Context, id string) (*models.InvestmentModel, error) {
	return s.repo.GetInvestmentModel(ctx, id)
}

// GetModelLeaderboard returns the model leaderboard sorted by avg mean payout.
func (s *Service) GetModelLeaderboard(ctx context.Context) ([]models.LabLeaderboardEntry, error) {
	return s.repo.GetModelLeaderboard(ctx)
}

// ListEntries returns entries matching the filter.
func (s *Service) ListEntries(ctx context.Context, filter models.LabListEntriesFilter, page models.LabPagination) ([]models.LabEntryDetail, error) {
	page.Limit, page.Offset = clampPagination(page.Limit, page.Offset)
	return s.repo.ListEntries(ctx, filter, page)
}

// GetEntryEnriched returns a single entry with enriched bids (team names, seeds, rational allocation).
func (s *Service) GetEntryEnriched(ctx context.Context, id string) (*models.LabEntryDetailEnriched, error) {
	raw, err := s.repo.GetEntryRaw(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting raw entry: %w", err)
	}
	return EnrichEntry(raw), nil
}

// GetEntryEnrichedByModelAndCalcutta returns an enriched entry for a model/calcutta pair.
func (s *Service) GetEntryEnrichedByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (*models.LabEntryDetailEnriched, error) {
	entryID, err := s.repo.GetEntryIDByModelAndCalcutta(ctx, modelName, calcuttaID, startingStateKey)
	if err != nil {
		return nil, fmt.Errorf("getting entry id for model: %w", err)
	}
	enriched, err := s.GetEntryEnriched(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("getting enriched entry: %w", err)
	}
	return enriched, nil
}

// ListEvaluations returns evaluations matching the filter.
func (s *Service) ListEvaluations(ctx context.Context, filter models.LabListEvaluationsFilter, page models.LabPagination) ([]models.LabEvaluationDetail, error) {
	page.Limit, page.Offset = clampPagination(page.Limit, page.Offset)
	return s.repo.ListEvaluations(ctx, filter, page)
}

// GetEvaluation returns a single evaluation by ID with full details.
func (s *Service) GetEvaluation(ctx context.Context, id string) (*models.LabEvaluationDetail, error) {
	return s.repo.GetEvaluation(ctx, id)
}

// GetEvaluationEntryResults returns per-entry results for an evaluation.
func (s *Service) GetEvaluationEntryResults(ctx context.Context, evaluationID string) ([]models.LabEvaluationEntryResult, error) {
	return s.repo.GetEvaluationEntryResults(ctx, evaluationID)
}

// GetEvaluationEntryProfile returns detailed profile for an entry result.
func (s *Service) GetEvaluationEntryProfile(ctx context.Context, entryResultID string) (*models.LabEvaluationEntryProfile, error) {
	return s.repo.GetEvaluationEntryProfile(ctx, entryResultID)
}

// GetEvaluationSummary returns the summary for an evaluation. If the summary
// has not been computed yet, it generates it on-the-fly, persists it, and
// returns it.
func (s *Service) GetEvaluationSummary(ctx context.Context, evaluationID string) (*models.LabEvaluationSummary, error) {
	// Check if summary already exists
	raw, err := s.repo.GetEvaluationSummaryJSON(ctx, evaluationID)
	if err != nil {
		return nil, fmt.Errorf("checking existing summary: %w", err)
	}
	if raw != nil {
		var summary models.LabEvaluationSummary
		if err := json.Unmarshal(raw, &summary); err != nil {
			return nil, fmt.Errorf("unmarshalling existing summary: %w", err)
		}
		return &summary, nil
	}

	// Generate on-the-fly
	return s.generateAndPersistSummary(ctx, evaluationID)
}

// generateAndPersistSummary builds a new summary, persists it, and returns it.
func (s *Service) generateAndPersistSummary(ctx context.Context, evaluationID string) (*models.LabEvaluationSummary, error) {
	eval, err := s.repo.GetEvaluation(ctx, evaluationID)
	if err != nil {
		return nil, fmt.Errorf("getting evaluation: %w", err)
	}

	entryResults, err := s.repo.GetEvaluationEntryResults(ctx, evaluationID)
	if err != nil {
		return nil, fmt.Errorf("getting entry results: %w", err)
	}

	// Find "Our Strategy" entry result to get bids
	var ourBids []models.LabEvaluationEntryBid
	for _, er := range entryResults {
		if er.EntryName == models.LabStrategyEntryName {
			profile, profileErr := s.repo.GetEvaluationEntryProfile(ctx, er.ID)
			if profileErr == nil && profile != nil {
				ourBids = profile.Bids
			}
			break
		}
	}

	// Try to find baseline (naive_ev) evaluation for comparison
	baselineEval, _ := s.repo.GetBaselineEvaluation(ctx, eval.CalcuttaID, eval.StartingStateKey)

	summary := BuildEvaluationSummary(entryResults, ourBids, baselineEval)

	// Persist
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("marshalling summary: %w", err)
	}
	if err := s.repo.UpdateEvaluationSummary(ctx, evaluationID, summaryBytes); err != nil {
		return nil, fmt.Errorf("persisting summary: %w", err)
	}

	return summary, nil
}

// StartPipeline starts a new pipeline run for a model.
func (s *Service) StartPipeline(ctx context.Context, modelID string, req models.LabStartPipelineRequest) (*models.LabStartPipelineResponse, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}

	// If force_rerun, delete existing artifacts first (this also cancels active pipelines)
	if req.ForceRerun {
		if err := s.pipelineRepo.SoftDeleteModelArtifacts(ctx, modelID); err != nil {
			return nil, fmt.Errorf("failed to clear existing artifacts: %w", err)
		}
	} else {
		// Only check for active pipeline if not force re-running
		active, err := s.pipelineRepo.GetActivePipelineRun(ctx, modelID)
		if err != nil {
			return nil, fmt.Errorf("checking active pipeline: %w", err)
		}
		if active != nil {
			return nil, &PipelineAlreadyRunningError{PipelineRunID: active.ID}
		}
	}

	// Get target calcutta IDs
	calcuttaIDs := req.CalcuttaIDs
	if len(calcuttaIDs) == 0 {
		var err error
		calcuttaIDs, err = s.pipelineRepo.GetHistoricalCalcuttaIDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting historical calcutta ids: %w", err)
		}
	}
	if len(calcuttaIDs) == 0 {
		return nil, &NoCalcuttasAvailableError{}
	}

	// Set defaults for optional parameters
	budgetPoints := req.BudgetPoints
	if budgetPoints <= 0 {
		budgetPoints = 100
	}
	optimizerKind := req.OptimizerKind
	if optimizerKind == "" {
		optimizerKind = "predicted_market_share"
	}
	nSims := req.NSims
	if nSims <= 0 {
		nSims = s.cfg.DefaultNSims
		if nSims <= 0 {
			nSims = 10000
		}
	}
	seed := req.Seed
	if seed == 0 {
		seed = 42
	}

	// Determine excluded entry name (request value takes precedence over config)
	excludedEntryName := req.ExcludedEntryName
	if excludedEntryName == "" {
		excludedEntryName = s.cfg.ExcludedEntryName
	}

	// Create pipeline run
	run := &models.LabPipelineRun{
		InvestmentModelID: modelID,
		TargetCalcuttaIDs: calcuttaIDs,
		BudgetPoints:      budgetPoints,
		OptimizerKind:     optimizerKind,
		NSims:             nSims,
		Seed:              seed,
		Status:            "pending",
	}
	if excludedEntryName != "" {
		run.ExcludedEntryName = &excludedEntryName
	}

	created, err := s.pipelineRepo.CreatePipelineRun(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("creating pipeline run: %w", err)
	}

	// Create calcutta runs
	err = s.pipelineRepo.CreatePipelineCalcuttaRuns(ctx, created.ID, calcuttaIDs)
	if err != nil {
		return nil, fmt.Errorf("creating pipeline calcutta runs: %w", err)
	}

	return &models.LabStartPipelineResponse{
		PipelineRunID: created.ID,
		NCalcuttas:    len(calcuttaIDs),
		Status:        "pending",
	}, nil
}

// GetPipelineProgress returns the progress for a specific pipeline run.
func (s *Service) GetPipelineProgress(ctx context.Context, pipelineRunID string) (*models.LabPipelineProgressResponse, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetPipelineProgress(ctx, pipelineRunID)
}

// GetModelPipelineProgress returns the pipeline progress for a model.
func (s *Service) GetModelPipelineProgress(ctx context.Context, modelID string) (*models.LabModelPipelineProgress, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetModelPipelineProgress(ctx, modelID)
}

// GetPipelineRun returns a pipeline run by ID.
func (s *Service) GetPipelineRun(ctx context.Context, id string) (*models.LabPipelineRun, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetPipelineRun(ctx, id)
}

// CancelPipeline cancels a running pipeline.
func (s *Service) CancelPipeline(ctx context.Context, pipelineRunID string) error {
	if s.pipelineRepo == nil {
		return &PipelineNotAvailableError{}
	}

	run, err := s.pipelineRepo.GetPipelineRun(ctx, pipelineRunID)
	if err != nil {
		return fmt.Errorf("getting pipeline run: %w", err)
	}

	if run.Status != "pending" && run.Status != "running" {
		return &PipelineNotCancellableError{Status: run.Status}
	}

	msg := "cancelled by user"
	if err := s.pipelineRepo.UpdatePipelineRunStatus(ctx, pipelineRunID, "cancelled", &msg); err != nil {
		return fmt.Errorf("updating pipeline run status: %w", err)
	}
	return nil
}

// Pipeline errors

// PipelineNotAvailableError indicates pipeline functionality is not available.
type PipelineNotAvailableError struct{}

func (e *PipelineNotAvailableError) Error() string {
	return "pipeline functionality not available"
}

// PipelineAlreadyRunningError indicates a pipeline is already running for the model.
type PipelineAlreadyRunningError struct {
	PipelineRunID string
}

func (e *PipelineAlreadyRunningError) Error() string {
	return "pipeline already running: " + e.PipelineRunID
}

// NoCalcuttasAvailableError indicates no historical calcuttas are available.
type NoCalcuttasAvailableError struct{}

func (e *NoCalcuttasAvailableError) Error() string {
	return "no historical calcuttas available"
}

// PipelineNotCancellableError indicates the pipeline cannot be cancelled.
type PipelineNotCancellableError struct {
	Status string
}

func (e *PipelineNotCancellableError) Error() string {
	return "pipeline cannot be cancelled: status is " + e.Status
}
