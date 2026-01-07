package httpserver

import "time"

type createSuiteCalcuttaEvaluationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type suiteCalcuttaEvaluationListItem struct {
	ID                        string     `json:"id"`
	SuiteExecutionID          *string    `json:"suite_execution_id,omitempty"`
	SuiteID                   string     `json:"suite_id"`
	SuiteName                 string     `json:"suite_name"`
	OptimizerKey              string     `json:"optimizer_key"`
	NSims                     int        `json:"n_sims"`
	Seed                      int        `json:"seed"`
	OurRank                   *int       `json:"our_rank,omitempty"`
	OurMeanNormalizedPayout   *float64   `json:"our_mean_normalized_payout,omitempty"`
	OurMedianNormalizedPayout *float64   `json:"our_median_normalized_payout,omitempty"`
	OurPTop1                  *float64   `json:"our_p_top1,omitempty"`
	OurPInMoney               *float64   `json:"our_p_in_money,omitempty"`
	TotalSimulations          *int       `json:"total_simulations,omitempty"`
	CalcuttaID                string     `json:"calcutta_id"`
	GameOutcomeRunID          *string    `json:"game_outcome_run_id,omitempty"`
	MarketShareRunID          *string    `json:"market_share_run_id,omitempty"`
	StrategyGenerationRunID   *string    `json:"strategy_generation_run_id,omitempty"`
	CalcuttaEvaluationRunID   *string    `json:"calcutta_evaluation_run_id,omitempty"`
	RealizedFinishPosition    *int       `json:"realized_finish_position,omitempty"`
	RealizedIsTied            *bool      `json:"realized_is_tied,omitempty"`
	RealizedInTheMoney        *bool      `json:"realized_in_the_money,omitempty"`
	RealizedPayoutCents       *int       `json:"realized_payout_cents,omitempty"`
	RealizedTotalPoints       *float64   `json:"realized_total_points,omitempty"`
	StartingStateKey          string     `json:"starting_state_key"`
	ExcludedEntryName         *string    `json:"excluded_entry_name,omitempty"`
	Status                    string     `json:"status"`
	ClaimedAt                 *time.Time `json:"claimed_at,omitempty"`
	ClaimedBy                 *string    `json:"claimed_by,omitempty"`
	ErrorMessage              *string    `json:"error_message,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

type suiteCalcuttaEvaluationListResponse struct {
	Items []suiteCalcuttaEvaluationListItem `json:"items"`
}

type suiteCalcuttaEvaluationPortfolioBid struct {
	TeamID      string  `json:"team_id"`
	SchoolName  string  `json:"school_name"`
	Seed        int     `json:"seed"`
	Region      string  `json:"region"`
	BidPoints   int     `json:"bid_points"`
	ExpectedROI float64 `json:"expected_roi"`
}

type suiteCalcuttaEvaluationOurStrategyPerformance struct {
	Rank                   int     `json:"rank"`
	EntryName              string  `json:"entry_name"`
	MeanNormalizedPayout   float64 `json:"mean_normalized_payout"`
	MedianNormalizedPayout float64 `json:"median_normalized_payout"`
	PTop1                  float64 `json:"p_top1"`
	PInMoney               float64 `json:"p_in_money"`
	TotalSimulations       int     `json:"total_simulations"`
}

type suiteCalcuttaEvaluationEntryPerformance struct {
	Rank                 int      `json:"rank"`
	EntryName            string   `json:"entry_name"`
	SnapshotEntryID      *string  `json:"snapshot_entry_id,omitempty"`
	MeanNormalizedPayout float64  `json:"mean_normalized_payout"`
	PTop1                float64  `json:"p_top1"`
	PInMoney             float64  `json:"p_in_money"`
	FinishPosition       *int     `json:"finish_position,omitempty"`
	IsTied               *bool    `json:"is_tied,omitempty"`
	InTheMoney           *bool    `json:"in_the_money,omitempty"`
	PayoutCents          *int     `json:"payout_cents,omitempty"`
	TotalPoints          *float64 `json:"total_points,omitempty"`
}

type suiteCalcuttaSnapshotEntryTeam struct {
	TeamID    string `json:"team_id"`
	School    string `json:"school_name"`
	Seed      int    `json:"seed"`
	Region    string `json:"region"`
	BidPoints int    `json:"bid_points"`
}

type suiteCalcuttaSnapshotEntryResponse struct {
	SnapshotEntryID string                           `json:"snapshot_entry_id"`
	DisplayName     string                           `json:"display_name"`
	IsSynthetic     bool                             `json:"is_synthetic"`
	Teams           []suiteCalcuttaSnapshotEntryTeam `json:"teams"`
}

type suiteCalcuttaEvaluationResultResponse struct {
	Evaluation  suiteCalcuttaEvaluationListItem                `json:"evaluation"`
	Portfolio   []suiteCalcuttaEvaluationPortfolioBid          `json:"portfolio"`
	OurStrategy *suiteCalcuttaEvaluationOurStrategyPerformance `json:"our_strategy,omitempty"`
	Entries     []suiteCalcuttaEvaluationEntryPerformance      `json:"entries"`
}

type suiteCalcuttaEvalFinish struct {
	FinishPosition int
	IsTied         bool
	InTheMoney     bool
	PayoutCents    int
	TotalPoints    float64
}
