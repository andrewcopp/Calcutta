package lab

import (
	"context"
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

// NewWithPipelineRepo creates a new lab service with pipeline repository support.
func NewWithPipelineRepo(repo ports.LabPipelineRepository, cfg ServiceConfig) *Service {
	return &Service{repo: repo, pipelineRepo: repo, cfg: cfg}
}

// ListInvestmentModels returns investment models matching the filter.
func (s *Service) ListInvestmentModels(ctx context.Context, filter models.LabListModelsFilter, page models.LabPagination) ([]models.InvestmentModel, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
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
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListEntries(ctx, filter, page)
}

// GetEntryEnriched returns a single entry with enriched bids (team names, seeds, naive allocation).
func (s *Service) GetEntryEnriched(ctx context.Context, id string) (*models.LabEntryDetailEnriched, error) {
	return s.repo.GetEntryEnriched(ctx, id)
}

// GetEntryEnrichedByModelAndCalcutta returns an enriched entry for a model/calcutta pair.
func (s *Service) GetEntryEnrichedByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (*models.LabEntryDetailEnriched, error) {
	return s.repo.GetEntryEnrichedByModelAndCalcutta(ctx, modelName, calcuttaID, startingStateKey)
}

// ListEvaluations returns evaluations matching the filter.
func (s *Service) ListEvaluations(ctx context.Context, filter models.LabListEvaluationsFilter, page models.LabPagination) ([]models.LabEvaluationDetail, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
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
			return nil, err
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
			return nil, err
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
		return nil, err
	}

	// Create calcutta runs
	err = s.pipelineRepo.CreatePipelineCalcuttaRuns(ctx, created.ID, calcuttaIDs)
	if err != nil {
		return nil, err
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
		return err
	}

	if run.Status != "pending" && run.Status != "running" {
		return &PipelineNotCancellableError{Status: run.Status}
	}

	msg := "cancelled by user"
	return s.pipelineRepo.UpdatePipelineRunStatus(ctx, pipelineRunID, "cancelled", &msg)
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
