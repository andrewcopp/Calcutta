package lab

import "time"

// PipelineRun represents a lab.pipeline_runs row.
type PipelineRun struct {
	ID                  string     `json:"id"`
	InvestmentModelID   string     `json:"investment_model_id"`
	TargetCalcuttaIDs   []string   `json:"target_calcutta_ids"`
	BudgetPoints        int        `json:"budget_points"`
	OptimizerKind       string     `json:"optimizer_kind"`
	NSims               int        `json:"n_sims"`
	Seed                int        `json:"seed"`
	ExcludedEntryName   *string    `json:"excluded_entry_name,omitempty"`
	Status              string     `json:"status"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	FinishedAt          *time.Time `json:"finished_at,omitempty"`
	ErrorMessage        *string    `json:"error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PipelineCalcuttaRun represents a lab.pipeline_calcutta_runs row.
type PipelineCalcuttaRun struct {
	ID                string     `json:"id"`
	PipelineRunID     string     `json:"pipeline_run_id"`
	CalcuttaID        string     `json:"calcutta_id"`
	EntryID           *string    `json:"entry_id,omitempty"`
	Stage             string     `json:"stage"`
	Status            string     `json:"status"`
	Progress          float64    `json:"progress"`
	ProgressMessage   *string    `json:"progress_message,omitempty"`
	PredictionsJobID  *string    `json:"predictions_job_id,omitempty"`
	OptimizationJobID *string    `json:"optimization_job_id,omitempty"`
	EvaluationJobID   *string    `json:"evaluation_job_id,omitempty"`
	EvaluationID      *string    `json:"evaluation_id,omitempty"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// StartPipelineRequest is the input for starting a pipeline run.
type StartPipelineRequest struct {
	CalcuttaIDs       []string `json:"calcutta_ids,omitempty"`
	BudgetPoints      int      `json:"budget_points,omitempty"`
	OptimizerKind     string   `json:"optimizer_kind,omitempty"`
	NSims             int      `json:"n_sims,omitempty"`
	Seed              int      `json:"seed,omitempty"`
	ExcludedEntryName string   `json:"excluded_entry_name,omitempty"`
	ForceRerun        bool     `json:"force_rerun,omitempty"`
}

// StartPipelineResponse is returned when a pipeline is started.
type StartPipelineResponse struct {
	PipelineRunID string `json:"pipeline_run_id"`
	NCalcuttas    int    `json:"n_calcuttas"`
	Status        string `json:"status"`
}

// CalcuttaProgressResponse shows progress for a single calcutta in a pipeline.
type CalcuttaProgressResponse struct {
	CalcuttaID       string   `json:"calcutta_id"`
	CalcuttaName     string   `json:"calcutta_name"`
	CalcuttaYear     int      `json:"calcutta_year"`
	Stage            string   `json:"stage"`
	Status           string   `json:"status"`
	Progress         float64  `json:"progress"`
	ProgressMessage  *string  `json:"progress_message,omitempty"`
	HasPredictions   bool     `json:"has_predictions"`
	HasEntry         bool     `json:"has_entry"`
	HasEvaluation    bool     `json:"has_evaluation"`
	EntryID          *string  `json:"entry_id,omitempty"`
	EvaluationID     *string  `json:"evaluation_id,omitempty"`
	MeanPayout       *float64 `json:"mean_payout,omitempty"`
	ErrorMessage     *string  `json:"error_message,omitempty"`
}

// PipelineProgressSummary shows aggregate progress across all calcuttas.
type PipelineProgressSummary struct {
	TotalCalcuttas    int      `json:"total_calcuttas"`
	PredictionsCount  int      `json:"predictions_count"`
	EntriesCount      int      `json:"entries_count"`
	EvaluationsCount  int      `json:"evaluations_count"`
	FailedCount       int      `json:"failed_count"`
	AvgMeanPayout     *float64 `json:"avg_mean_payout,omitempty"`
}

// PipelineProgressResponse is the full progress response for a pipeline run.
type PipelineProgressResponse struct {
	ID               string                     `json:"id"`
	InvestmentModelID string                    `json:"investment_model_id"`
	ModelName        string                     `json:"model_name"`
	Status           string                     `json:"status"`
	StartedAt        *time.Time                 `json:"started_at,omitempty"`
	FinishedAt       *time.Time                 `json:"finished_at,omitempty"`
	ErrorMessage     *string                    `json:"error_message,omitempty"`
	Summary          PipelineProgressSummary    `json:"summary"`
	Calcuttas        []CalcuttaProgressResponse `json:"calcuttas"`
}

// ModelPipelineProgress shows the current pipeline state for a model.
// This is used even when no pipeline is running, to show existing artifacts.
type ModelPipelineProgress struct {
	ModelID              string                     `json:"model_id"`
	ModelName            string                     `json:"model_name"`
	ActivePipelineRunID  *string                    `json:"active_pipeline_run_id,omitempty"`
	TotalCalcuttas       int                        `json:"total_calcuttas"`
	PredictionsCount     int                        `json:"predictions_count"`
	EntriesCount         int                        `json:"entries_count"`
	EvaluationsCount     int                        `json:"evaluations_count"`
	AvgMeanPayout        *float64                   `json:"avg_mean_payout,omitempty"`
	Calcuttas            []CalcuttaProgressResponse `json:"calcuttas"`
}

// PipelineRepository extends Repository with pipeline-specific methods.
type PipelineRepository interface {
	Repository

	// Pipeline run operations
	CreatePipelineRun(run *PipelineRun) (*PipelineRun, error)
	GetPipelineRun(id string) (*PipelineRun, error)
	UpdatePipelineRunStatus(id string, status string, errorMessage *string) error
	ListPipelineRuns(modelID *string, status *string, limit int) ([]PipelineRun, error)
	GetActivePipelineRun(modelID string) (*PipelineRun, error)

	// Pipeline calcutta run operations
	CreatePipelineCalcuttaRuns(pipelineRunID string, calcuttaIDs []string) error
	GetPipelineCalcuttaRuns(pipelineRunID string) ([]PipelineCalcuttaRun, error)
	UpdatePipelineCalcuttaRun(id string, updates map[string]interface{}) error

	// Progress queries
	GetPipelineProgress(pipelineRunID string) (*PipelineProgressResponse, error)
	GetModelPipelineProgress(modelID string) (*ModelPipelineProgress, error)

	// Historical calcuttas for pipeline
	GetHistoricalCalcuttaIDs() ([]string, error)

	// Cleanup for force re-run
	SoftDeleteModelArtifacts(modelID string) error
}
