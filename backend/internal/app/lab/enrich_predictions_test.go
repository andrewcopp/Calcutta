package lab

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ---------------------------------------------------------------------------
// helpers (prediction-specific)
// ---------------------------------------------------------------------------

func findPredictionByTeamID(preds []models.LabEnrichedPrediction, teamID string) *models.LabEnrichedPrediction {
	for i := range preds {
		if preds[i].TeamID == teamID {
			return &preds[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// enrichPredictions
// ---------------------------------------------------------------------------

func TestThatEnrichPredictionsComputesPredictedBidPointsFromMarketShare(t *testing.T) {
	// GIVEN a prediction with 25% market share and pool budget of 4000
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN predicted bid points = 0.25 * 4000 = 1000
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.PredictedBidPoints != 1000 {
		t.Errorf("expected predicted bid points 1000, got %d", pred.PredictedBidPoints)
	}
}

func TestThatEnrichPredictionsComputesExpectedROI(t *testing.T) {
	// GIVEN a prediction with expected points 320 and predicted bid points 1000
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN expected ROI = 320 / 1000 = 0.32
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	expected := 320.0 / 1000.0
	if !floatEquals(pred.ExpectedROI, expected, epsilon) {
		t.Errorf("expected ROI %f, got %f", expected, pred.ExpectedROI)
	}
}

func TestThatEnrichPredictionsExpectedROIIsZeroWhenPredictedBidPointsIsZero(t *testing.T) {
	// GIVEN a prediction with 0% market share (predicted bid points = 0)
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.0, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN expected ROI is zero (avoids divide by zero)
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.ExpectedROI != 0.0 {
		t.Errorf("expected ROI 0.0 when predicted bid points is zero, got %f", pred.ExpectedROI)
	}
}

func TestThatEnrichPredictionsComputesNaivePointsUsingTotalPoolBudget(t *testing.T) {
	// GIVEN two teams with a pool budget of 10000 and team-a has 75% of expected points
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 10000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 300.0},
			{TeamID: "team-b", PredictedMarketShare: 0.3, ExpectedPoints: 100.0},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN team-a naive points = 300/400 * 10000 = 7500
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.NaivePoints != 7500 {
		t.Errorf("expected naive points 7500, got %d", pred.NaivePoints)
	}
}

func TestThatEnrichPredictionsComputesPositiveEdgeForUndervaluedTeam(t *testing.T) {
	// GIVEN a team with naive points 7500 and predicted bid points 5000
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 10000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 300.0},
			{TeamID: "team-b", PredictedMarketShare: 0.3, ExpectedPoints: 100.0},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN team-a has positive edge (naive 7500 > predicted 5000)
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	// edge = (7500 - 5000) / 7500 * 100 = 33.33...
	if pred.EdgePercent <= 0 {
		t.Errorf("expected positive edge percent for undervalued team, got %f", pred.EdgePercent)
	}
}

func TestThatEnrichPredictionsComputesNegativeEdgeForOvervaluedTeam(t *testing.T) {
	// GIVEN a team with naive points 2500 and predicted bid points 3000
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 10000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 300.0},
			{TeamID: "team-b", PredictedMarketShare: 0.3, ExpectedPoints: 100.0},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN team-b has negative edge (naive 2500 < predicted 3000)
	pred := findPredictionByTeamID(result, "team-b")
	if pred == nil {
		t.Fatal("expected prediction for team-b")
	}
	if pred.EdgePercent >= 0 {
		t.Errorf("expected negative edge percent for overvalued team, got %f", pred.EdgePercent)
	}
}

func TestThatEnrichPredictionsComputesCorrectEdgePercentValue(t *testing.T) {
	// GIVEN a team where naive = 7500 and predicted bid = 5000
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 10000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 300.0},
			{TeamID: "team-b", PredictedMarketShare: 0.3, ExpectedPoints: 100.0},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN edge = (7500 - 5000) / 7500 * 100 = 33.333...
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	expected := float64(7500-5000) / float64(7500) * 100
	if !floatEquals(pred.EdgePercent, expected, epsilon) {
		t.Errorf("expected edge percent %f, got %f", expected, pred.EdgePercent)
	}
}

func TestThatEnrichPredictionsPreservesMarketShare(t *testing.T) {
	// GIVEN a prediction with 25% market share
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN the market share value is preserved
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.PredictedMarketShare != 0.25 {
		t.Errorf("expected market share 0.25, got %f", pred.PredictedMarketShare)
	}
}

func TestThatEnrichPredictionsPreservesExpectedPoints(t *testing.T) {
	// GIVEN a prediction with 320 expected points
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN the expected points value is preserved
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.ExpectedPoints != 320.0 {
		t.Errorf("expected 320.0, got %f", pred.ExpectedPoints)
	}
}

func TestThatEnrichPredictionsPreservesSchoolName(t *testing.T) {
	// GIVEN a team named "Duke"
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN the school name is preserved
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.SchoolName != "Duke" {
		t.Errorf("expected 'Duke', got '%s'", pred.SchoolName)
	}
}

func TestThatEnrichPredictionsSkipsTeamNotInTeamMap(t *testing.T) {
	// GIVEN a prediction for a team that does not appear in the Teams map
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 320.0},
			{TeamID: "team-unknown", PredictedMarketShare: 0.1, ExpectedPoints: 50.0},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0, "team-unknown": 50.0}
	totalEP := 370.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN only the team in the team map is included
	if len(result) != 1 {
		t.Errorf("expected 1 prediction (skipping unknown team), got %d", len(result))
	}
}

func TestThatEnrichPredictionsEdgePercentIsZeroWhenNaivePointsIsZero(t *testing.T) {
	// GIVEN a team with zero expected points (naive points will be 0)
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.25, ExpectedPoints: 0.0},
		},
	}
	teamEP := map[string]float64{"team-a": 0.0}
	totalEP := 100.0 // nonzero to avoid NaN in naiveShare

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN edge percent is zero (avoids divide by zero)
	pred := findPredictionByTeamID(result, "team-a")
	if pred == nil {
		t.Fatal("expected prediction for team-a")
	}
	if pred.EdgePercent != 0.0 {
		t.Errorf("expected edge percent 0.0, got %f", pred.EdgePercent)
	}
}

func TestThatEnrichPredictionsReturnsEmptySliceWhenNoPredictions(t *testing.T) {
	// GIVEN no predictions
	raw := &models.LabEntryRaw{
		TotalPoolBudget: 4000,
		Teams:           twoTeamMap(),
		Predictions:     []models.LabPrediction{},
	}
	teamEP := map[string]float64{"team-a": 320.0, "team-b": 1.0}
	totalEP := 321.0

	// WHEN enriching predictions
	result := enrichPredictions(raw, teamEP, totalEP)

	// THEN result is empty
	if len(result) != 0 {
		t.Errorf("expected 0 predictions, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// EnrichEntry -- prediction-focused behavior
// ---------------------------------------------------------------------------

func TestThatEnrichEntryReturnsNilPredictionsWhenHasPredictionsIsFalse(t *testing.T) {
	// GIVEN an entry without predictions
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          twoTeamMap(),
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 500},
			{TeamID: "team-b", BidPoints: 500},
		},
		Predictions: []models.LabPrediction{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN predictions is nil
	if result.Predictions != nil {
		t.Errorf("expected nil predictions when HasPredictions is false, got %d", len(result.Predictions))
	}
}

func TestThatEnrichEntryPopulatesPredictionsWhenHasPredictionsIsTrue(t *testing.T) {
	// GIVEN an entry with predictions
	raw := &models.LabEntryRaw{
		ID:              "entry-1",
		HasPredictions:  true,
		TotalPoolBudget: 4000,
		Teams:           twoTeamMap(),
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 500},
			{TeamID: "team-b", BidPoints: 500},
		},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.8, ExpectedPoints: 320.0},
			{TeamID: "team-b", PredictedMarketShare: 0.2, ExpectedPoints: 1.0},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN predictions slice is populated
	if len(result.Predictions) != 2 {
		t.Errorf("expected 2 predictions, got %d", len(result.Predictions))
	}
}
