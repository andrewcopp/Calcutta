package lab

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ---------------------------------------------------------------------------
// sumValues
// ---------------------------------------------------------------------------

func TestThatSumValuesReturnsZeroForEmptyMap(t *testing.T) {
	// GIVEN an empty map
	m := map[string]float64{}

	// WHEN summing values
	result := sumValues(m)

	// THEN the result is zero
	if result != 0.0 {
		t.Errorf("expected 0.0, got %f", result)
	}
}

func TestThatSumValuesReturnsSingleEntryValue(t *testing.T) {
	// GIVEN a map with one entry
	m := map[string]float64{"a": 42.5}

	// WHEN summing values
	result := sumValues(m)

	// THEN the result equals the single value
	if result != 42.5 {
		t.Errorf("expected 42.5, got %f", result)
	}
}

func TestThatSumValuesReturnsSumOfAllPositiveValues(t *testing.T) {
	// GIVEN a map with multiple positive values
	m := map[string]float64{"a": 10.0, "b": 20.0, "c": 30.0}

	// WHEN summing values
	result := sumValues(m)

	// THEN the result is the sum of all values
	if result != 60.0 {
		t.Errorf("expected 60.0, got %f", result)
	}
}

func TestThatSumValuesIncludesNegativeValues(t *testing.T) {
	// GIVEN a map with positive and negative values
	m := map[string]float64{"a": 100.0, "b": -25.0}

	// WHEN summing values
	result := sumValues(m)

	// THEN the result includes the negative value
	if result != 75.0 {
		t.Errorf("expected 75.0, got %f", result)
	}
}

// ---------------------------------------------------------------------------
// buildTeamExpectedPoints
// ---------------------------------------------------------------------------

func TestThatBuildTeamExpectedPointsUsesPredictionValuesWhenAvailable(t *testing.T) {
	// GIVEN a raw entry with predictions containing expected points
	raw := &models.LabEntryRaw{
		HasPredictions: true,
		Teams:          twoTeamMap(),
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", ExpectedPoints: 500.0, PredictedMarketShare: 0.5},
			{TeamID: "team-b", ExpectedPoints: 10.0, PredictedMarketShare: 0.1},
		},
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN team-a uses the prediction expected points of 500
	if result["team-a"] != 500.0 {
		t.Errorf("expected 500.0 for team-a, got %f", result["team-a"])
	}
}

func TestThatBuildTeamExpectedPointsUsesAllPredictionValues(t *testing.T) {
	// GIVEN a raw entry with predictions for two teams
	raw := &models.LabEntryRaw{
		HasPredictions: true,
		Teams:          twoTeamMap(),
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", ExpectedPoints: 500.0, PredictedMarketShare: 0.5},
			{TeamID: "team-b", ExpectedPoints: 10.0, PredictedMarketShare: 0.1},
		},
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN team-b uses the prediction expected points of 10
	if result["team-b"] != 10.0 {
		t.Errorf("expected 10.0 for team-b, got %f", result["team-b"])
	}
}

func TestThatBuildTeamExpectedPointsFallsBackToSeedWhenNoPredictions(t *testing.T) {
	// GIVEN a raw entry without predictions for a 1-seed team
	raw := &models.LabEntryRaw{
		HasPredictions: false,
		Teams:          twoTeamMap(), // team-a is seed 1, team-b is seed 16
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN team-a (seed 1) uses the seed-based fallback of 320
	if result["team-a"] != SeedExpectedPoints[1] {
		t.Errorf("expected %f for seed 1, got %f", SeedExpectedPoints[1], result["team-a"])
	}
}

func TestThatBuildTeamExpectedPointsSeedFallbackUsesCorrectValueForSixteenSeed(t *testing.T) {
	// GIVEN a raw entry without predictions for a 16-seed team
	raw := &models.LabEntryRaw{
		HasPredictions: false,
		Teams:          twoTeamMap(), // team-b is seed 16
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN team-b (seed 16) uses the seed-based fallback of 1
	if result["team-b"] != SeedExpectedPoints[16] {
		t.Errorf("expected %f for seed 16, got %f", SeedExpectedPoints[16], result["team-b"])
	}
}

func TestThatBuildTeamExpectedPointsReturnsEntryForEachTeam(t *testing.T) {
	// GIVEN a raw entry without predictions and three teams
	raw := &models.LabEntryRaw{
		HasPredictions: false,
		Teams:          threeTeamMap(),
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN the map has one entry per team
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
}

func TestThatBuildTeamExpectedPointsReturnsEmptyMapWhenNoTeams(t *testing.T) {
	// GIVEN a raw entry with no teams and no predictions
	raw := &models.LabEntryRaw{
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
	}

	// WHEN building team expected points
	result := buildTeamExpectedPoints(raw)

	// THEN the map is empty
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

// ---------------------------------------------------------------------------
// EnrichEntry -- expected points / seed fallback behavior
// ---------------------------------------------------------------------------

func TestThatEnrichEntryUsessSeedFallbackWhenNoPredictions(t *testing.T) {
	// GIVEN an entry without predictions for a seed-1 and seed-16 team
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          twoTeamMap(), // team-a seed 1, team-b seed 16
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 500},
			{TeamID: "team-b", BidPoints: 500},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN team-a (seed 1) rational allocation is proportional to seed 1 expected points
	// seed 1 = 320, seed 16 = 1, total = 321
	// rational for team-a = 320/321 * 1000 = 996 (int truncation)
	bid := findBidByTeamID(result.Bids, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	expectedRational := int(SeedExpectedPoints[1] / (SeedExpectedPoints[1] + SeedExpectedPoints[16]) * 1000.0)
	if bid.RationalPoints != expectedRational {
		t.Errorf("expected rational points %d, got %d", expectedRational, bid.RationalPoints)
	}
}

func TestThatEnrichEntryUsesPredictionExpectedPointsForRationalAllocation(t *testing.T) {
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
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 600.0},
			{TeamID: "team-b", PredictedMarketShare: 0.5, ExpectedPoints: 400.0},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN bid rational allocation uses prediction expected points (600/(600+400) * 1000 = 600)
	bid := findBidByTeamID(result.Bids, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	expectedRational := int(600.0 / 1000.0 * 1000.0)
	if bid.RationalPoints != expectedRational {
		t.Errorf("expected rational points %d, got %d", expectedRational, bid.RationalPoints)
	}
}
