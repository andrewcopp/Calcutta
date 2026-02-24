package models

import (
	"encoding/json"
	"time"
)

// LabStrategyEntryName is the canonical display name for the lab-generated entry
// when it participates in simulated calcutta evaluations.
const LabStrategyEntryName = "Our Strategy"

// InvestmentModel represents a lab.investment_models row.
type InvestmentModel struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Kind         string          `json:"kind"`
	ParamsJSON   json.RawMessage `json:"paramsJson"`
	Notes        *string         `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	NEntries     int             `json:"nEntries"`
	NEvaluations int             `json:"nEvaluations"`
}

// LabEntry represents a lab.entries row.
type LabEntry struct {
	ID                    string          `json:"id"`
	InvestmentModelID     string          `json:"investmentModelId"`
	CalcuttaID            string          `json:"calcuttaId"`
	GameOutcomeKind       string          `json:"gameOutcomeKind"`
	GameOutcomeParamsJSON json.RawMessage `json:"gameOutcomeParamsJson"`
	OptimizerKind         string          `json:"optimizerKind"`
	OptimizerParamsJSON   json.RawMessage `json:"optimizerParamsJson"`
	StartingStateKey      string          `json:"startingStateKey"`
	PredictionsJSON       json.RawMessage `json:"predictionsJson,omitempty"`
	BidsJSON              json.RawMessage `json:"bidsJson"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

// LabEntryDetail is LabEntry with joined data for display.
type LabEntryDetail struct {
	LabEntry
	ModelName    string `json:"modelName"`
	ModelKind    string `json:"modelKind"`
	CalcuttaName string `json:"calcuttaName"`
	NEvaluations int    `json:"nEvaluations"`
}

// LabPrediction represents a single market prediction within predictions_json.
type LabPrediction struct {
	TeamID               string  `json:"teamId"`
	PredictedMarketShare float64 `json:"predictedMarketShare"`
	ExpectedPoints       float64 `json:"expectedPoints"`
}

// LabEnrichedPrediction is LabPrediction with team details for display.
type LabEnrichedPrediction struct {
	TeamID               string  `json:"teamId"`
	SchoolName           string  `json:"schoolName"`
	Seed                 int     `json:"seed"`
	Region               string  `json:"region"`
	PredictedMarketShare float64 `json:"predictedMarketShare"`
	PredictedBidPoints   int     `json:"predictedBidPoints"`
	ExpectedPoints       float64 `json:"expectedPoints"`
	ExpectedROI          float64 `json:"expectedRoi"`
	RationalPoints       int     `json:"rationalPoints"`
	EdgePercent          float64 `json:"edgePercent"`
}

// LabEntryBid represents a single optimized bid within bids_json.
type LabEntryBid struct {
	TeamID      string   `json:"teamId"`
	BidPoints   int      `json:"bidPoints"`
	ExpectedROI *float64 `json:"expectedRoi,omitempty"`
}

// LabEnrichedBid is LabEntryBid with team details and rational allocation for display.
type LabEnrichedBid struct {
	TeamID         string   `json:"teamId"`
	SchoolName     string   `json:"schoolName"`
	Seed           int      `json:"seed"`
	Region         string   `json:"region"`
	BidPoints      int      `json:"bidPoints"`
	RationalPoints int      `json:"rationalPoints"`
	EdgePercent    float64  `json:"edgePercent"`
	ExpectedROI    *float64 `json:"expectedRoi,omitempty"`
}

// LabTeamInfo holds team metadata used during enrichment.
type LabTeamInfo struct {
	Name   string
	Seed   int
	Region string
}

// LabEntryRaw holds the raw entry data fetched from the database before enrichment.
type LabEntryRaw struct {
	ID                    string
	InvestmentModelID     string
	CalcuttaID            string
	GameOutcomeKind       string
	GameOutcomeParamsJSON json.RawMessage
	OptimizerKind         string
	OptimizerParamsJSON   json.RawMessage
	StartingStateKey      string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ModelName             string
	ModelKind             string
	CalcuttaName          string
	NEvaluations          int
	HasPredictions        bool
	Predictions           []LabPrediction
	Bids                  []LabEntryBid
	Teams                 map[string]LabTeamInfo
	TotalPoolBudget       int
}

// LabEntryDetailEnriched is LabEntryDetail with enriched predictions and bids.
type LabEntryDetailEnriched struct {
	ID                    string                 `json:"id"`
	InvestmentModelID     string                 `json:"investmentModelId"`
	CalcuttaID            string                 `json:"calcuttaId"`
	GameOutcomeKind       string                 `json:"gameOutcomeKind"`
	GameOutcomeParamsJSON json.RawMessage        `json:"gameOutcomeParamsJson"`
	OptimizerKind         string                 `json:"optimizerKind"`
	OptimizerParamsJSON   json.RawMessage        `json:"optimizerParamsJson"`
	StartingStateKey      string                 `json:"startingStateKey"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
	ModelName             string                 `json:"modelName"`
	ModelKind             string                 `json:"modelKind"`
	CalcuttaName          string                 `json:"calcuttaName"`
	NEvaluations          int                    `json:"nEvaluations"`
	HasPredictions        bool                   `json:"hasPredictions"`
	Predictions           []LabEnrichedPrediction `json:"predictions,omitempty"`
	Bids                  []LabEnrichedBid        `json:"bids"`
}

// LabEvaluation represents a lab.evaluations row.
type LabEvaluation struct {
	ID                     string          `json:"id"`
	EntryID                string          `json:"entryId"`
	NSims                  int             `json:"nSims"`
	Seed                   int             `json:"seed"`
	MeanNormalizedPayout   *float64        `json:"meanNormalizedPayout,omitempty"`
	MedianNormalizedPayout *float64        `json:"medianNormalizedPayout,omitempty"`
	PTop1                  *float64        `json:"pTop1,omitempty"`
	PInMoney               *float64        `json:"pInMoney,omitempty"`
	OurRank                *int            `json:"ourRank,omitempty"`
	SummaryJSON            json.RawMessage `json:"summaryJson,omitempty"`
	CreatedAt              time.Time       `json:"createdAt"`
	UpdatedAt              time.Time       `json:"updatedAt"`
}

// LabEvaluationSummary is the structured narrative computed for an evaluation.
type LabEvaluationSummary struct {
	PercentileRank float64                       `json:"percentileRank"`
	VsBaseline     *LabEvaluationBaselineComparison `json:"vsBaseline,omitempty"`
	NEntries       int                           `json:"nEntries"`
	TopHoldings    []LabEvaluationTopHolding     `json:"topHoldings"`
	KeyInsight     string                        `json:"keyInsight"`
}

// LabEvaluationBaselineComparison compares an evaluation to a naive_ev baseline.
type LabEvaluationBaselineComparison struct {
	MeanPayoutDelta float64 `json:"meanPayoutDelta"`
	PTop1Delta      float64 `json:"pTop1Delta"`
	Interpretation  string  `json:"interpretation"`
}

// LabEvaluationTopHolding represents a top bid by points in the strategy.
type LabEvaluationTopHolding struct {
	SchoolName string `json:"schoolName"`
	Seed       int    `json:"seed"`
	BidPoints  int    `json:"bidPoints"`
}

// LabEvaluationDetail is LabEvaluation with joined data for display.
type LabEvaluationDetail struct {
	LabEvaluation
	ModelName        string `json:"modelName"`
	ModelKind        string `json:"modelKind"`
	CalcuttaID       string `json:"calcuttaId"`
	CalcuttaName     string `json:"calcuttaName"`
	StartingStateKey string `json:"startingStateKey"`
}

// LabEvaluationEntryResult represents a single entry's result in an evaluation.
type LabEvaluationEntryResult struct {
	ID                   string   `json:"id"`
	EntryName            string   `json:"entryName"`
	MeanNormalizedPayout *float64 `json:"meanNormalizedPayout,omitempty"`
	PTop1                *float64 `json:"pTop1,omitempty"`
	PInMoney             *float64 `json:"pInMoney,omitempty"`
	Rank                 int      `json:"rank"`
}

// LabEvaluationEntryBid represents a single team bid within an evaluation entry.
type LabEvaluationEntryBid struct {
	TeamID     string  `json:"teamId"`
	SchoolName string  `json:"schoolName"`
	Seed       int     `json:"seed"`
	Region     string  `json:"region"`
	BidPoints  int     `json:"bidPoints"`
	Ownership  float64 `json:"ownership"`
}

// LabEvaluationEntryProfile represents a full entry profile with result and bids.
type LabEvaluationEntryProfile struct {
	EntryName            string                  `json:"entryName"`
	MeanNormalizedPayout *float64                `json:"meanNormalizedPayout,omitempty"`
	PTop1                *float64                `json:"pTop1,omitempty"`
	PInMoney             *float64                `json:"pInMoney,omitempty"`
	Rank                 int                     `json:"rank"`
	TotalBidPoints       int                     `json:"totalBidPoints"`
	Bids                 []LabEvaluationEntryBid  `json:"bids"`
}

// LabLeaderboardEntry represents a row from lab.model_leaderboard view.
type LabLeaderboardEntry struct {
	InvestmentModelID         string     `json:"investmentModelId"`
	ModelName                 string     `json:"modelName"`
	ModelKind                 string     `json:"modelKind"`
	NEntries                  int        `json:"nEntries"`
	NEntriesWithPredictions   int        `json:"nEntriesWithPredictions"`
	NEvaluations              int        `json:"nEvaluations"`
	NCalcuttasWithEntries     int        `json:"nCalcuttasWithEntries"`
	NCalcuttasWithEvaluations int        `json:"nCalcuttasWithEvaluations"`
	AvgMeanPayout             *float64   `json:"avgMeanPayout,omitempty"`
	AvgMedianPayout           *float64   `json:"avgMedianPayout,omitempty"`
	AvgPTop1                  *float64   `json:"avgPTop1,omitempty"`
	AvgPInMoney               *float64   `json:"avgPInMoney,omitempty"`
	FirstEvalAt               *time.Time `json:"firstEvalAt,omitempty"`
	LastEvalAt                *time.Time `json:"lastEvalAt,omitempty"`
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

