package testutil

import (
	"encoding/json"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// NewLabPrediction returns a fully-populated LabPrediction with sensible defaults.
func NewLabPrediction() models.LabPrediction {
	return models.LabPrediction{
		TeamID:               "team-1",
		PredictedMarketShare: 0.05,
		ExpectedPoints:       10.0,
	}
}

// NewLabEntryBid returns a fully-populated LabEntryBid with sensible defaults.
func NewLabEntryBid() models.LabEntryBid {
	return models.LabEntryBid{
		TeamID:    "team-1",
		BidPoints: 10,
	}
}

// NewLabTeamInfo returns a fully-populated LabTeamInfo with sensible defaults.
func NewLabTeamInfo() models.LabTeamInfo {
	return models.LabTeamInfo{
		Name:   "Test School",
		Seed:   1,
		Region: "East",
	}
}

// NewLabEntryRaw returns a fully-populated LabEntryRaw with sensible defaults.
func NewLabEntryRaw() *models.LabEntryRaw {
	return &models.LabEntryRaw{
		ID:                    "lab-entry-1",
		InvestmentModelID:     "model-1",
		CalcuttaID:            "calcutta-1",
		GameOutcomeKind:       "kenpom",
		GameOutcomeParamsJSON: json.RawMessage(`{}`),
		OptimizerKind:         "greedy",
		OptimizerParamsJSON:   json.RawMessage(`{}`),
		StartingStateKey:      "default",
		CreatedAt:             DefaultTime,
		UpdatedAt:             DefaultTime,
		ModelName:             "Test Model",
		ModelKind:             "test",
		CalcuttaName:          "Test Calcutta",
		Predictions:           []models.LabPrediction{},
		Bids:                  []models.LabEntryBid{},
		Teams:                 map[string]models.LabTeamInfo{},
		TotalPoolBudget:       100,
	}
}
