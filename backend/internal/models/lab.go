package models

import (
	"encoding/json"
	"time"
)

// InvestmentModel represents a lab.investment_models row.
type InvestmentModel struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Kind         string          `json:"kind"`
	ParamsJSON   json.RawMessage `json:"params_json"`
	Notes        *string         `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	NEntries     int             `json:"n_entries"`
	NEvaluations int             `json:"n_evaluations"`
}

// LabEntry represents a lab.entries row.
type LabEntry struct {
	ID                    string          `json:"id"`
	InvestmentModelID     string          `json:"investment_model_id"`
	CalcuttaID            string          `json:"calcutta_id"`
	GameOutcomeKind       string          `json:"game_outcome_kind"`
	GameOutcomeParamsJSON json.RawMessage `json:"game_outcome_params_json"`
	OptimizerKind         string          `json:"optimizer_kind"`
	OptimizerParamsJSON   json.RawMessage `json:"optimizer_params_json"`
	StartingStateKey      string          `json:"starting_state_key"`
	PredictionsJSON       json.RawMessage `json:"predictions_json,omitempty"`
	BidsJSON              json.RawMessage `json:"bids_json"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// LabEntryDetail is LabEntry with joined data for display.
type LabEntryDetail struct {
	LabEntry
	ModelName    string `json:"model_name"`
	ModelKind    string `json:"model_kind"`
	CalcuttaName string `json:"calcutta_name"`
	NEvaluations int    `json:"n_evaluations"`
}

// LabPrediction represents a single market prediction within predictions_json.
type LabPrediction struct {
	TeamID               string  `json:"team_id"`
	PredictedMarketShare float64 `json:"predicted_market_share"`
	ExpectedPoints       float64 `json:"expected_points"`
}

// LabEnrichedPrediction is LabPrediction with team details for display.
type LabEnrichedPrediction struct {
	TeamID               string  `json:"team_id"`
	SchoolName           string  `json:"school_name"`
	Seed                 int     `json:"seed"`
	Region               string  `json:"region"`
	PredictedMarketShare float64 `json:"predicted_market_share"`
	PredictedBidPoints   int     `json:"predicted_bid_points"`
	ExpectedPoints       float64 `json:"expected_points"`
	ExpectedROI          float64 `json:"expected_roi"`
	NaivePoints          int     `json:"naive_points"`
	EdgePercent          float64 `json:"edge_percent"`
}

// LabEntryBid represents a single optimized bid within bids_json.
type LabEntryBid struct {
	TeamID      string   `json:"team_id"`
	BidPoints   int      `json:"bid_points"`
	ExpectedROI *float64 `json:"expected_roi,omitempty"`
}

// LabEnrichedBid is LabEntryBid with team details and naive allocation for display.
type LabEnrichedBid struct {
	TeamID      string   `json:"team_id"`
	SchoolName  string   `json:"school_name"`
	Seed        int      `json:"seed"`
	Region      string   `json:"region"`
	BidPoints   int      `json:"bid_points"`
	NaivePoints int      `json:"naive_points"`
	EdgePercent float64  `json:"edge_percent"`
	ExpectedROI *float64 `json:"expected_roi,omitempty"`
}

// LabEntryDetailEnriched is LabEntryDetail with enriched predictions and bids.
type LabEntryDetailEnriched struct {
	ID                    string                 `json:"id"`
	InvestmentModelID     string                 `json:"investment_model_id"`
	CalcuttaID            string                 `json:"calcutta_id"`
	GameOutcomeKind       string                 `json:"game_outcome_kind"`
	GameOutcomeParamsJSON json.RawMessage        `json:"game_outcome_params_json"`
	OptimizerKind         string                 `json:"optimizer_kind"`
	OptimizerParamsJSON   json.RawMessage        `json:"optimizer_params_json"`
	StartingStateKey      string                 `json:"starting_state_key"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	ModelName             string                 `json:"model_name"`
	ModelKind             string                 `json:"model_kind"`
	CalcuttaName          string                 `json:"calcutta_name"`
	NEvaluations          int                    `json:"n_evaluations"`
	HasPredictions        bool                   `json:"has_predictions"`
	Predictions           []LabEnrichedPrediction `json:"predictions,omitempty"`
	Bids                  []LabEnrichedBid        `json:"bids"`
}

// LabEvaluation represents a lab.evaluations row.
type LabEvaluation struct {
	ID                     string    `json:"id"`
	EntryID                string    `json:"entry_id"`
	NSims                  int       `json:"n_sims"`
	Seed                   int       `json:"seed"`
	MeanNormalizedPayout   *float64  `json:"mean_normalized_payout,omitempty"`
	MedianNormalizedPayout *float64  `json:"median_normalized_payout,omitempty"`
	PTop1                  *float64  `json:"p_top1,omitempty"`
	PInMoney               *float64  `json:"p_in_money,omitempty"`
	OurRank                *int      `json:"our_rank,omitempty"`
	SimulatedCalcuttaID    *string   `json:"simulated_calcutta_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// LabEvaluationDetail is LabEvaluation with joined data for display.
type LabEvaluationDetail struct {
	LabEvaluation
	ModelName        string `json:"model_name"`
	ModelKind        string `json:"model_kind"`
	CalcuttaID       string `json:"calcutta_id"`
	CalcuttaName     string `json:"calcutta_name"`
	StartingStateKey string `json:"starting_state_key"`
}

// LabEvaluationEntryResult represents a single entry's result in an evaluation.
type LabEvaluationEntryResult struct {
	ID                   string   `json:"id"`
	EntryName            string   `json:"entry_name"`
	MeanNormalizedPayout *float64 `json:"mean_normalized_payout,omitempty"`
	PTop1                *float64 `json:"p_top1,omitempty"`
	PInMoney             *float64 `json:"p_in_money,omitempty"`
	Rank                 int      `json:"rank"`
}

// LabEvaluationEntryBid represents a single team bid within an evaluation entry.
type LabEvaluationEntryBid struct {
	TeamID     string  `json:"team_id"`
	SchoolName string  `json:"school_name"`
	Seed       int     `json:"seed"`
	Region     string  `json:"region"`
	BidPoints  int     `json:"bid_points"`
	Ownership  float64 `json:"ownership"`
}

// LabEvaluationEntryProfile represents a full entry profile with result and bids.
type LabEvaluationEntryProfile struct {
	EntryName            string                  `json:"entry_name"`
	MeanNormalizedPayout *float64                `json:"mean_normalized_payout,omitempty"`
	PTop1                *float64                `json:"p_top1,omitempty"`
	PInMoney             *float64                `json:"p_in_money,omitempty"`
	Rank                 int                     `json:"rank"`
	TotalBidPoints       int                     `json:"total_bid_points"`
	Bids                 []LabEvaluationEntryBid  `json:"bids"`
}

// LabLeaderboardEntry represents a row from lab.model_leaderboard view.
type LabLeaderboardEntry struct {
	InvestmentModelID         string     `json:"investment_model_id"`
	ModelName                 string     `json:"model_name"`
	ModelKind                 string     `json:"model_kind"`
	NEntries                  int        `json:"n_entries"`
	NEntriesWithPredictions   int        `json:"n_entries_with_predictions"`
	NEvaluations              int        `json:"n_evaluations"`
	NCalcuttasWithEntries     int        `json:"n_calcuttas_with_entries"`
	NCalcuttasWithEvaluations int        `json:"n_calcuttas_with_evaluations"`
	AvgMeanPayout             *float64   `json:"avg_mean_payout,omitempty"`
	AvgMedianPayout           *float64   `json:"avg_median_payout,omitempty"`
	AvgPTop1                  *float64   `json:"avg_p_top1,omitempty"`
	AvgPInMoney               *float64   `json:"avg_p_in_money,omitempty"`
	FirstEvalAt               *time.Time `json:"first_eval_at,omitempty"`
	LastEvalAt                *time.Time `json:"last_eval_at,omitempty"`
}

// LabListModelsFilter for filtering investment models list.
type LabListModelsFilter struct {
	Kind *string
}

// LabListEntriesFilter for filtering entries list.
type LabListEntriesFilter struct {
	InvestmentModelID *string
	CalcuttaID        *string
	StartingStateKey  *string
}

// LabListEvaluationsFilter for filtering evaluations list.
type LabListEvaluationsFilter struct {
	EntryID           *string
	InvestmentModelID *string
	CalcuttaID        *string
}

// LabPagination for list operations.
type LabPagination struct {
	Limit  int
	Offset int
}

// LabGenerateEntriesRequest for triggering entry generation via Python.
type LabGenerateEntriesRequest struct {
	Years         []int  `json:"years,omitempty"`
	BudgetPoints  int    `json:"budget_points,omitempty"`
	ExcludedEntry string `json:"excluded_entry,omitempty"`
}

// LabGenerateEntriesResponse from the Python entry generation script.
type LabGenerateEntriesResponse struct {
	EntriesCreated int      `json:"entries_created"`
	Errors         []string `json:"errors,omitempty"`
}
