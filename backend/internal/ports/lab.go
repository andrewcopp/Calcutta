package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// LabRepository defines the interface for lab data access.
type LabRepository interface {
	ListInvestmentModels(ctx context.Context, filter models.LabListModelsFilter, page models.LabPagination) ([]models.InvestmentModel, error)
	GetInvestmentModel(ctx context.Context, id string) (*models.InvestmentModel, error)
	GetModelLeaderboard(ctx context.Context) ([]models.LabLeaderboardEntry, error)
	ListEntries(ctx context.Context, filter models.LabListEntriesFilter, page models.LabPagination) ([]models.LabEntryDetail, error)
	GetEntryRaw(ctx context.Context, id string) (*models.LabEntryRaw, error)
	GetEntryIDByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (string, error)
	ListEvaluations(ctx context.Context, filter models.LabListEvaluationsFilter, page models.LabPagination) ([]models.LabEvaluationDetail, error)
	GetEvaluation(ctx context.Context, id string) (*models.LabEvaluationDetail, error)
	GetEvaluationEntryResults(ctx context.Context, evaluationID string) ([]models.LabEvaluationEntryResult, error)
	GetEvaluationEntryProfile(ctx context.Context, entryResultID string) (*models.LabEvaluationEntryProfile, error)
	UpdateEvaluationSummary(ctx context.Context, evaluationID string, summaryJSON []byte) error
	GetBaselineEvaluation(ctx context.Context, calcuttaID, startingStateKey string) (*models.LabEvaluationDetail, error)
	GetEvaluationSummaryJSON(ctx context.Context, evaluationID string) ([]byte, error)
}

// LabPipelineRepository extends LabRepository with pipeline-specific methods.
type LabPipelineRepository interface {
	LabRepository

	// Pipeline run operations
	CreatePipelineRun(ctx context.Context, run *models.LabPipelineRun) (*models.LabPipelineRun, error)
	GetPipelineRun(ctx context.Context, id string) (*models.LabPipelineRun, error)
	UpdatePipelineRunStatus(ctx context.Context, id string, status string, errorMessage *string) error
	GetActivePipelineRun(ctx context.Context, modelID string) (*models.LabPipelineRun, error)

	// Pipeline calcutta run operations
	CreatePipelineCalcuttaRuns(ctx context.Context, pipelineRunID string, calcuttaIDs []string) error

	// Progress queries
	GetPipelineProgress(ctx context.Context, pipelineRunID string) (*models.LabPipelineProgressResponse, error)
	GetModelPipelineProgress(ctx context.Context, modelID string) (*models.LabModelPipelineProgress, error)

	// Historical calcuttas for pipeline
	GetHistoricalCalcuttaIDs(ctx context.Context) ([]string, error)

	// Cleanup for force re-run
	SoftDeleteModelArtifacts(ctx context.Context, modelID string) error
}
