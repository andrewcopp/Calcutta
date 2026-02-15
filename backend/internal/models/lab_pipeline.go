package models

import "time"

// LabPipelineRun represents a lab.pipeline_runs row.
type LabPipelineRun struct {
	ID                string     `json:"id"`
	InvestmentModelID string     `json:"investment_model_id"`
	TargetCalcuttaIDs []string   `json:"target_calcutta_ids"`
	BudgetPoints      int        `json:"budget_points"`
	OptimizerKind     string     `json:"optimizer_kind"`
	NSims             int        `json:"n_sims"`
	Seed              int        `json:"seed"`
	ExcludedEntryName *string    `json:"excluded_entry_name,omitempty"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// LabPipelineCalcuttaRun represents a lab.pipeline_calcutta_runs row.
type LabPipelineCalcuttaRun struct {
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

// LabStartPipelineRequest is the input for starting a pipeline run.
type LabStartPipelineRequest struct {
	CalcuttaIDs       []string `json:"calcutta_ids,omitempty"`
	BudgetPoints      int      `json:"budget_points,omitempty"`
	OptimizerKind     string   `json:"optimizer_kind,omitempty"`
	NSims             int      `json:"n_sims,omitempty"`
	Seed              int      `json:"seed,omitempty"`
	ExcludedEntryName string   `json:"excluded_entry_name,omitempty"`
	ForceRerun        bool     `json:"force_rerun,omitempty"`
}

// LabStartPipelineResponse is returned when a pipeline is started.
type LabStartPipelineResponse struct {
	PipelineRunID string `json:"pipeline_run_id"`
	NCalcuttas    int    `json:"n_calcuttas"`
	Status        string `json:"status"`
}

// LabCalcuttaProgressResponse shows progress for a single calcutta in a pipeline.
type LabCalcuttaProgressResponse struct {
	CalcuttaID      string   `json:"calcutta_id"`
	CalcuttaName    string   `json:"calcutta_name"`
	CalcuttaYear    int      `json:"calcutta_year"`
	Stage           string   `json:"stage"`
	Status          string   `json:"status"`
	Progress        float64  `json:"progress"`
	ProgressMessage *string  `json:"progress_message,omitempty"`
	HasPredictions  bool     `json:"has_predictions"`
	HasEntry        bool     `json:"has_entry"`
	HasEvaluation   bool     `json:"has_evaluation"`
	EntryID         *string  `json:"entry_id,omitempty"`
	EvaluationID    *string  `json:"evaluation_id,omitempty"`
	MeanPayout      *float64 `json:"mean_payout,omitempty"`
	OurRank         *int     `json:"our_rank,omitempty"`
	ErrorMessage    *string  `json:"error_message,omitempty"`
}

// LabPipelineProgressSummary shows aggregate progress across all calcuttas.
type LabPipelineProgressSummary struct {
	TotalCalcuttas   int      `json:"total_calcuttas"`
	PredictionsCount int      `json:"predictions_count"`
	EntriesCount     int      `json:"entries_count"`
	EvaluationsCount int      `json:"evaluations_count"`
	FailedCount      int      `json:"failed_count"`
	AvgMeanPayout    *float64 `json:"avg_mean_payout,omitempty"`
}

// LabPipelineProgressResponse is the full progress response for a pipeline run.
type LabPipelineProgressResponse struct {
	ID                string                       `json:"id"`
	InvestmentModelID string                       `json:"investment_model_id"`
	ModelName         string                       `json:"model_name"`
	Status            string                       `json:"status"`
	StartedAt         *time.Time                   `json:"started_at,omitempty"`
	FinishedAt        *time.Time                   `json:"finished_at,omitempty"`
	ErrorMessage      *string                      `json:"error_message,omitempty"`
	Summary           LabPipelineProgressSummary    `json:"summary"`
	Calcuttas         []LabCalcuttaProgressResponse `json:"calcuttas"`
}

// LabModelPipelineProgress shows the current pipeline state for a model.
type LabModelPipelineProgress struct {
	ModelID             string                       `json:"model_id"`
	ModelName           string                       `json:"model_name"`
	ActivePipelineRunID *string                      `json:"active_pipeline_run_id,omitempty"`
	TotalCalcuttas      int                          `json:"total_calcuttas"`
	PredictionsCount    int                          `json:"predictions_count"`
	EntriesCount        int                          `json:"entries_count"`
	EvaluationsCount    int                          `json:"evaluations_count"`
	AvgMeanPayout       *float64                     `json:"avg_mean_payout,omitempty"`
	Calcuttas           []LabCalcuttaProgressResponse `json:"calcuttas"`
}
