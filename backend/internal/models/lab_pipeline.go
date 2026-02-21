package models

import "time"

// LabPipelineRun represents a lab.pipeline_runs row.
type LabPipelineRun struct {
	ID                string     `json:"id"`
	InvestmentModelID string     `json:"investmentModelId"`
	TargetCalcuttaIDs []string   `json:"targetCalcuttaIds"`
	BudgetPoints      int        `json:"budgetPoints"`
	OptimizerKind     string     `json:"optimizerKind"`
	NSims             int        `json:"nSims"`
	Seed              int        `json:"seed"`
	ExcludedEntryName *string    `json:"excludedEntryName,omitempty"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"startedAt,omitempty"`
	FinishedAt        *time.Time `json:"finishedAt,omitempty"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// LabStartPipelineRequest is the input for starting a pipeline run.
type LabStartPipelineRequest struct {
	CalcuttaIDs       []string `json:"calcuttaIds,omitempty"`
	BudgetPoints      int      `json:"budgetPoints,omitempty"`
	OptimizerKind     string   `json:"optimizerKind,omitempty"`
	NSims             int      `json:"nSims,omitempty"`
	Seed              int      `json:"seed,omitempty"`
	ExcludedEntryName string   `json:"excludedEntryName,omitempty"`
	ForceRerun        bool     `json:"forceRerun,omitempty"`
}

// LabStartPipelineResponse is returned when a pipeline is started.
type LabStartPipelineResponse struct {
	PipelineRunID string `json:"pipelineRunId"`
	NCalcuttas    int    `json:"nCalcuttas"`
	Status        string `json:"status"`
}

// LabCalcuttaProgressResponse shows progress for a single calcutta in a pipeline.
type LabCalcuttaProgressResponse struct {
	CalcuttaID      string   `json:"calcuttaId"`
	CalcuttaName    string   `json:"calcuttaName"`
	CalcuttaYear    int      `json:"calcuttaYear"`
	Stage           string   `json:"stage"`
	Status          string   `json:"status"`
	Progress        float64  `json:"progress"`
	ProgressMessage *string  `json:"progressMessage,omitempty"`
	HasPredictions  bool     `json:"hasPredictions"`
	HasEntry        bool     `json:"hasEntry"`
	HasEvaluation   bool     `json:"hasEvaluation"`
	EntryID         *string  `json:"entryId,omitempty"`
	EvaluationID    *string  `json:"evaluationId,omitempty"`
	MeanPayout      *float64 `json:"meanPayout,omitempty"`
	OurRank         *int     `json:"ourRank,omitempty"`
	ErrorMessage    *string  `json:"errorMessage,omitempty"`
}

// LabPipelineProgressSummary shows aggregate progress across all calcuttas.
type LabPipelineProgressSummary struct {
	TotalCalcuttas   int      `json:"totalCalcuttas"`
	PredictionsCount int      `json:"predictionsCount"`
	EntriesCount     int      `json:"entriesCount"`
	EvaluationsCount int      `json:"evaluationsCount"`
	FailedCount      int      `json:"failedCount"`
	AvgMeanPayout    *float64 `json:"avgMeanPayout,omitempty"`
}

// LabPipelineProgressResponse is the full progress response for a pipeline run.
type LabPipelineProgressResponse struct {
	ID                string                       `json:"id"`
	InvestmentModelID string                       `json:"investmentModelId"`
	ModelName         string                       `json:"modelName"`
	Status            string                       `json:"status"`
	StartedAt         *time.Time                   `json:"startedAt,omitempty"`
	FinishedAt        *time.Time                   `json:"finishedAt,omitempty"`
	ErrorMessage      *string                      `json:"errorMessage,omitempty"`
	Summary           LabPipelineProgressSummary    `json:"summary"`
	Calcuttas         []LabCalcuttaProgressResponse `json:"calcuttas"`
}

// LabModelPipelineProgress shows the current pipeline state for a model.
type LabModelPipelineProgress struct {
	ModelID             string                       `json:"modelId"`
	ModelName           string                       `json:"modelName"`
	ActivePipelineRunID *string                      `json:"activePipelineRunId,omitempty"`
	TotalCalcuttas      int                          `json:"totalCalcuttas"`
	PredictionsCount    int                          `json:"predictionsCount"`
	EntriesCount        int                          `json:"entriesCount"`
	EvaluationsCount    int                          `json:"evaluationsCount"`
	AvgMeanPayout       *float64                     `json:"avgMeanPayout,omitempty"`
	Calcuttas           []LabCalcuttaProgressResponse `json:"calcuttas"`
}
