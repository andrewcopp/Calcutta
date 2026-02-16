package ports

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// LabRepository defines the interface for lab data access.
type LabRepository interface {
	ListInvestmentModels(filter models.LabListModelsFilter, page models.LabPagination) ([]models.InvestmentModel, error)
	GetInvestmentModel(id string) (*models.InvestmentModel, error)
	GetModelLeaderboard() ([]models.LabLeaderboardEntry, error)
	ListEntries(filter models.LabListEntriesFilter, page models.LabPagination) ([]models.LabEntryDetail, error)
	GetEntryEnriched(id string) (*models.LabEntryDetailEnriched, error)
	GetEntryEnrichedByModelAndCalcutta(modelName, calcuttaID, startingStateKey string) (*models.LabEntryDetailEnriched, error)
	ListEvaluations(filter models.LabListEvaluationsFilter, page models.LabPagination) ([]models.LabEvaluationDetail, error)
	GetEvaluation(id string) (*models.LabEvaluationDetail, error)
	GetEvaluationEntryResults(evaluationID string) ([]models.LabEvaluationEntryResult, error)
	GetEvaluationEntryProfile(entryResultID string) (*models.LabEvaluationEntryProfile, error)
}

// LabPipelineRepository extends LabRepository with pipeline-specific methods.
type LabPipelineRepository interface {
	LabRepository

	// Pipeline run operations
	CreatePipelineRun(run *models.LabPipelineRun) (*models.LabPipelineRun, error)
	GetPipelineRun(id string) (*models.LabPipelineRun, error)
	UpdatePipelineRunStatus(id string, status string, errorMessage *string) error
	ListPipelineRuns(modelID *string, status *string, limit int) ([]models.LabPipelineRun, error)
	GetActivePipelineRun(modelID string) (*models.LabPipelineRun, error)

	// Pipeline calcutta run operations
	CreatePipelineCalcuttaRuns(pipelineRunID string, calcuttaIDs []string) error
	GetPipelineCalcuttaRuns(pipelineRunID string) ([]models.LabPipelineCalcuttaRun, error)
	UpdatePipelineCalcuttaRun(id string, updates map[string]interface{}) error

	// Progress queries
	GetPipelineProgress(pipelineRunID string) (*models.LabPipelineProgressResponse, error)
	GetModelPipelineProgress(modelID string) (*models.LabModelPipelineProgress, error)

	// Historical calcuttas for pipeline
	GetHistoricalCalcuttaIDs() ([]string, error)

	// Cleanup for force re-run
	SoftDeleteModelArtifacts(modelID string) error
}
