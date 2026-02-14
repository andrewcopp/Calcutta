package lab

import (
	"encoding/json"
	"time"
)

// InvestmentModel represents a lab.investment_models row.
type InvestmentModel struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Kind        string          `json:"kind"`
	ParamsJSON  json.RawMessage `json:"params_json"`
	Notes       *string         `json:"notes,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	NEntries    int             `json:"n_entries"`
	NEvaluations int            `json:"n_evaluations"`
}

// Entry represents a lab.entries row.
type Entry struct {
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

// EntryDetail is Entry with joined data for display.
type EntryDetail struct {
	Entry
	ModelName    string `json:"model_name"`
	ModelKind    string `json:"model_kind"`
	CalcuttaName string `json:"calcutta_name"`
	NEvaluations int    `json:"n_evaluations"`
}

// Prediction represents a single market prediction within predictions_json.
// This is what the model predicts THE MARKET will bid on a team.
type Prediction struct {
	TeamID               string  `json:"team_id"`
	PredictedMarketShare float64 `json:"predicted_market_share"`
	ExpectedPoints       float64 `json:"expected_points"`
}

// EnrichedPrediction is Prediction with team details for display.
type EnrichedPrediction struct {
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

// EntryBid represents a single optimized bid within bids_json.
// This is OUR optimal allocation given the market predictions.
type EntryBid struct {
	TeamID      string   `json:"team_id"`
	BidPoints   int      `json:"bid_points"`
	ExpectedROI *float64 `json:"expected_roi,omitempty"`
}

// EnrichedBid is EntryBid with team details and naive allocation for display.
type EnrichedBid struct {
	TeamID       string   `json:"team_id"`
	SchoolName   string   `json:"school_name"`
	Seed         int      `json:"seed"`
	Region       string   `json:"region"`
	BidPoints    int      `json:"bid_points"`
	NaivePoints  int      `json:"naive_points"`
	EdgePercent  float64  `json:"edge_percent"`
	ExpectedROI  *float64 `json:"expected_roi,omitempty"`
}

// EntryDetailEnriched is EntryDetail with enriched predictions and bids for display.
type EntryDetailEnriched struct {
	ID                    string               `json:"id"`
	InvestmentModelID     string               `json:"investment_model_id"`
	CalcuttaID            string               `json:"calcutta_id"`
	GameOutcomeKind       string               `json:"game_outcome_kind"`
	GameOutcomeParamsJSON json.RawMessage      `json:"game_outcome_params_json"`
	OptimizerKind         string               `json:"optimizer_kind"`
	OptimizerParamsJSON   json.RawMessage      `json:"optimizer_params_json"`
	StartingStateKey      string               `json:"starting_state_key"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	ModelName             string               `json:"model_name"`
	ModelKind             string               `json:"model_kind"`
	CalcuttaName          string               `json:"calcutta_name"`
	NEvaluations          int                  `json:"n_evaluations"`
	HasPredictions        bool                 `json:"has_predictions"`
	Predictions           []EnrichedPrediction `json:"predictions,omitempty"`
	Bids                  []EnrichedBid        `json:"bids"`
}

// Evaluation represents a lab.evaluations row.
type Evaluation struct {
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

// EvaluationDetail is Evaluation with joined data for display.
type EvaluationDetail struct {
	Evaluation
	ModelName        string `json:"model_name"`
	ModelKind        string `json:"model_kind"`
	CalcuttaID       string `json:"calcutta_id"`
	CalcuttaName     string `json:"calcutta_name"`
	StartingStateKey string `json:"starting_state_key"`
}

// EvaluationEntryResult represents a single entry's result in an evaluation.
type EvaluationEntryResult struct {
	ID                   string   `json:"id"`
	EntryName            string   `json:"entry_name"`
	MeanNormalizedPayout *float64 `json:"mean_normalized_payout,omitempty"`
	PTop1                *float64 `json:"p_top1,omitempty"`
	PInMoney             *float64 `json:"p_in_money,omitempty"`
	Rank                 int      `json:"rank"`
}

// EvaluationEntryBid represents a single team bid within an evaluation entry.
type EvaluationEntryBid struct {
	TeamID     string  `json:"team_id"`
	SchoolName string  `json:"school_name"`
	Seed       int     `json:"seed"`
	Region     string  `json:"region"`
	BidPoints  int     `json:"bid_points"`
	Ownership  float64 `json:"ownership"` // Percentage of total pool invested in this team
}

// EvaluationEntryProfile represents a full entry profile with result and bids.
type EvaluationEntryProfile struct {
	EntryName            string               `json:"entry_name"`
	MeanNormalizedPayout *float64             `json:"mean_normalized_payout,omitempty"`
	PTop1                *float64             `json:"p_top1,omitempty"`
	PInMoney             *float64             `json:"p_in_money,omitempty"`
	Rank                 int                  `json:"rank"`
	TotalBidPoints       int                  `json:"total_bid_points"`
	Bids                 []EvaluationEntryBid `json:"bids"`
}

// LeaderboardEntry represents a row from lab.model_leaderboard view.
type LeaderboardEntry struct {
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

// ListModelsFilter for filtering investment models list.
type ListModelsFilter struct {
	Kind *string
}

// ListEntriesFilter for filtering entries list.
type ListEntriesFilter struct {
	InvestmentModelID *string
	CalcuttaID        *string
	StartingStateKey  *string
}

// ListEvaluationsFilter for filtering evaluations list.
type ListEvaluationsFilter struct {
	EntryID           *string
	InvestmentModelID *string
	CalcuttaID        *string
}

// Pagination for list operations.
type Pagination struct {
	Limit  int
	Offset int
}

// GenerateEntriesRequest for triggering entry generation via Python.
type GenerateEntriesRequest struct {
	Years         []int  `json:"years,omitempty"`
	BudgetPoints  int    `json:"budget_points,omitempty"`
	ExcludedEntry string `json:"excluded_entry,omitempty"`
}

// GenerateEntriesResponse from the Python entry generation script.
type GenerateEntriesResponse struct {
	EntriesCreated int      `json:"entries_created"`
	Errors         []string `json:"errors,omitempty"`
}

// Repository defines the interface for lab data access.
type Repository interface {
	ListInvestmentModels(filter ListModelsFilter, page Pagination) ([]InvestmentModel, error)
	GetInvestmentModel(id string) (*InvestmentModel, error)
	GetModelLeaderboard() ([]LeaderboardEntry, error)
	ListEntries(filter ListEntriesFilter, page Pagination) ([]EntryDetail, error)
	GetEntry(id string) (*EntryDetail, error)
	GetEntryEnriched(id string) (*EntryDetailEnriched, error)
	GetEntryEnrichedByModelAndCalcutta(modelName, calcuttaID, startingStateKey string) (*EntryDetailEnriched, error)
	ListEvaluations(filter ListEvaluationsFilter, page Pagination) ([]EvaluationDetail, error)
	GetEvaluation(id string) (*EvaluationDetail, error)
	GetEvaluationEntryResults(evaluationID string) ([]EvaluationEntryResult, error)
	GetEvaluationEntryProfile(entryResultID string) (*EvaluationEntryProfile, error)
}
