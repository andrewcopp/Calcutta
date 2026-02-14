package lab

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

// Service provides lab-related business logic.
type Service struct {
	repo         Repository
	pipelineRepo PipelineRepository
}

// New creates a new lab service.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// NewWithPipelineRepo creates a new lab service with pipeline repository support.
func NewWithPipelineRepo(repo PipelineRepository) *Service {
	return &Service{repo: repo, pipelineRepo: repo}
}

// ListInvestmentModels returns investment models matching the filter.
func (s *Service) ListInvestmentModels(ctx context.Context, filter ListModelsFilter, page Pagination) ([]InvestmentModel, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListInvestmentModels(filter, page)
}

// GetInvestmentModel returns a single investment model by ID.
func (s *Service) GetInvestmentModel(ctx context.Context, id string) (*InvestmentModel, error) {
	return s.repo.GetInvestmentModel(id)
}

// GetModelLeaderboard returns the model leaderboard sorted by avg mean payout.
func (s *Service) GetModelLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	return s.repo.GetModelLeaderboard()
}

// ListEntries returns entries matching the filter.
func (s *Service) ListEntries(ctx context.Context, filter ListEntriesFilter, page Pagination) ([]EntryDetail, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListEntries(filter, page)
}

// GetEntry returns a single entry by ID with full details.
func (s *Service) GetEntry(ctx context.Context, id string) (*EntryDetail, error) {
	return s.repo.GetEntry(id)
}

// GetEntryEnriched returns a single entry with enriched bids (team names, seeds, naive allocation).
func (s *Service) GetEntryEnriched(ctx context.Context, id string) (*EntryDetailEnriched, error) {
	return s.repo.GetEntryEnriched(id)
}

// GetEntryEnrichedByModelAndCalcutta returns an enriched entry for a model/calcutta pair.
func (s *Service) GetEntryEnrichedByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (*EntryDetailEnriched, error) {
	return s.repo.GetEntryEnrichedByModelAndCalcutta(modelName, calcuttaID, startingStateKey)
}

// ListEvaluations returns evaluations matching the filter.
func (s *Service) ListEvaluations(ctx context.Context, filter ListEvaluationsFilter, page Pagination) ([]EvaluationDetail, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListEvaluations(filter, page)
}

// GetEvaluation returns a single evaluation by ID with full details.
func (s *Service) GetEvaluation(ctx context.Context, id string) (*EvaluationDetail, error) {
	return s.repo.GetEvaluation(id)
}

// GetEvaluationEntryResults returns per-entry results for an evaluation.
func (s *Service) GetEvaluationEntryResults(ctx context.Context, evaluationID string) ([]EvaluationEntryResult, error) {
	return s.repo.GetEvaluationEntryResults(evaluationID)
}

// GetEvaluationEntryProfile returns detailed profile for an entry result.
func (s *Service) GetEvaluationEntryProfile(ctx context.Context, entryResultID string) (*EvaluationEntryProfile, error) {
	return s.repo.GetEvaluationEntryProfile(entryResultID)
}

// GenerateEntries runs the Python script to generate entries for a model.
func (s *Service) GenerateEntries(ctx context.Context, modelID string, req GenerateEntriesRequest) (*GenerateEntriesResponse, error) {
	return RunGenerateEntriesScript(ctx, modelID, req)
}

// StartPipeline starts a new pipeline run for a model.
func (s *Service) StartPipeline(ctx context.Context, modelID string, req StartPipelineRequest) (*StartPipelineResponse, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}

	// If force_rerun, delete existing artifacts first (this also cancels active pipelines)
	if req.ForceRerun {
		if err := s.pipelineRepo.SoftDeleteModelArtifacts(modelID); err != nil {
			return nil, fmt.Errorf("failed to clear existing artifacts: %w", err)
		}
	} else {
		// Only check for active pipeline if not force re-running
		active, err := s.pipelineRepo.GetActivePipelineRun(modelID)
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
		calcuttaIDs, err = s.pipelineRepo.GetHistoricalCalcuttaIDs()
		if err != nil {
			return nil, err
		}
	}
	if len(calcuttaIDs) == 0 {
		return nil, &NoCalcuttasAvailableError{}
	}

	// Set defaults
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
		nSims = 10000 // fallback default
		if envNSims := os.Getenv("DEFAULT_N_SIMS"); envNSims != "" {
			if parsed, err := strconv.Atoi(envNSims); err == nil && parsed > 0 {
				nSims = parsed
			}
		}
	}
	seed := req.Seed
	if seed == 0 {
		seed = 42
	}

	// Determine excluded entry name (request value takes precedence over env var)
	excludedEntryName := req.ExcludedEntryName
	if excludedEntryName == "" {
		excludedEntryName = os.Getenv("EXCLUDED_ENTRY_NAME")
	}

	// Create pipeline run
	run := &PipelineRun{
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

	created, err := s.pipelineRepo.CreatePipelineRun(run)
	if err != nil {
		return nil, err
	}

	// Create calcutta runs
	err = s.pipelineRepo.CreatePipelineCalcuttaRuns(created.ID, calcuttaIDs)
	if err != nil {
		return nil, err
	}

	return &StartPipelineResponse{
		PipelineRunID: created.ID,
		NCalcuttas:    len(calcuttaIDs),
		Status:        "pending",
	}, nil
}

// GetPipelineProgress returns the progress for a specific pipeline run.
func (s *Service) GetPipelineProgress(ctx context.Context, pipelineRunID string) (*PipelineProgressResponse, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetPipelineProgress(pipelineRunID)
}

// GetModelPipelineProgress returns the pipeline progress for a model.
func (s *Service) GetModelPipelineProgress(ctx context.Context, modelID string) (*ModelPipelineProgress, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetModelPipelineProgress(modelID)
}

// GetPipelineRun returns a pipeline run by ID.
func (s *Service) GetPipelineRun(ctx context.Context, id string) (*PipelineRun, error) {
	if s.pipelineRepo == nil {
		return nil, &PipelineNotAvailableError{}
	}
	return s.pipelineRepo.GetPipelineRun(id)
}

// CancelPipeline cancels a running pipeline.
func (s *Service) CancelPipeline(ctx context.Context, pipelineRunID string) error {
	if s.pipelineRepo == nil {
		return &PipelineNotAvailableError{}
	}

	run, err := s.pipelineRepo.GetPipelineRun(pipelineRunID)
	if err != nil {
		return err
	}

	if run.Status != "pending" && run.Status != "running" {
		return &PipelineNotCancellableError{Status: run.Status}
	}

	msg := "cancelled by user"
	return s.pipelineRepo.UpdatePipelineRunStatus(pipelineRunID, "cancelled", &msg)
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
