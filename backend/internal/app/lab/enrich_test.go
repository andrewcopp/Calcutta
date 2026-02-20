package lab

import (
	"math"
	"sort"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func makeTeams(entries ...struct {
	id     string
	name   string
	seed   int
	region string
}) map[string]models.LabTeamInfo {
	m := make(map[string]models.LabTeamInfo, len(entries))
	for _, e := range entries {
		m[e.id] = models.LabTeamInfo{Name: e.name, Seed: e.seed, Region: e.region}
	}
	return m
}

type teamDef struct {
	id     string
	name   string
	seed   int
	region string
}

func twoTeamMap() map[string]models.LabTeamInfo {
	return map[string]models.LabTeamInfo{
		"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		"team-b": {Name: "UNC", Seed: 16, Region: "East"},
	}
}

func threeTeamMap() map[string]models.LabTeamInfo {
	return map[string]models.LabTeamInfo{
		"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		"team-c": {Name: "Kansas", Seed: 3, Region: "West"},
	}
}

func findBidByTeamID(bids []models.LabEnrichedBid, teamID string) *models.LabEnrichedBid {
	for i := range bids {
		if bids[i].TeamID == teamID {
			return &bids[i]
		}
	}
	return nil
}

func findPredictionByTeamID(preds []models.LabEnrichedPrediction, teamID string) *models.LabEnrichedPrediction {
	for i := range preds {
		if preds[i].TeamID == teamID {
			return &preds[i]
		}
	}
	return nil
}

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

const epsilon = 0.001

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
// enrichBids
// ---------------------------------------------------------------------------

func TestThatEnrichBidsComputesCorrectNaivePointsForSingleTeam(t *testing.T) {
	// GIVEN a single team with 100% of expected points and 1000 budget
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 1000},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN naive points for the only team equals the full budget
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.NaivePoints != 1000 {
		t.Errorf("expected naive points 1000, got %d", bid.NaivePoints)
	}
}

func TestThatEnrichBidsComputesProportionalNaivePoints(t *testing.T) {
	// GIVEN two teams where team-a has 75% of expected points and budget is 1000
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 800},
			{TeamID: "team-b", BidPoints: 200},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN team-a gets 75% of budget = 750 naive points
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.NaivePoints != 750 {
		t.Errorf("expected naive points 750, got %d", bid.NaivePoints)
	}
}

func TestThatEnrichBidsComputesPositiveEdgeForUndervaluedTeam(t *testing.T) {
	// GIVEN a team whose actual bid (200) is lower than its naive allocation (750)
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 200},
			{TeamID: "team-b", BidPoints: 800},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN team-a has positive edge percent (undervalued: naive 750 > bid 200)
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	// edge = (750 - 200) / 750 * 100 = 73.333...
	if bid.EdgePercent <= 0 {
		t.Errorf("expected positive edge percent for undervalued team, got %f", bid.EdgePercent)
	}
}

func TestThatEnrichBidsComputesNegativeEdgeForOvervaluedTeam(t *testing.T) {
	// GIVEN a team whose actual bid (800) is higher than its naive allocation (250)
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 200},
			{TeamID: "team-b", BidPoints: 800},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN team-b has negative edge percent (overvalued: naive 250 < bid 800)
	bid := findBidByTeamID(result, "team-b")
	if bid == nil {
		t.Fatal("expected bid for team-b")
	}
	// edge = (250 - 800) / 250 * 100 = -220
	if bid.EdgePercent >= 0 {
		t.Errorf("expected negative edge percent for overvalued team, got %f", bid.EdgePercent)
	}
}

func TestThatEnrichBidsComputesCorrectEdgePercentValue(t *testing.T) {
	// GIVEN a team with naive points 750 and bid 200
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 2, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 200},
			{TeamID: "team-b", BidPoints: 800},
		},
	}
	teamEP := map[string]float64{"team-a": 300.0, "team-b": 100.0}
	totalEP := 400.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN team-a edge = (750 - 200) / 750 * 100 = 73.333...
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	expected := float64(750-200) / float64(750) * 100
	if !floatEquals(bid.EdgePercent, expected, epsilon) {
		t.Errorf("expected edge percent %f, got %f", expected, bid.EdgePercent)
	}
}

func TestThatEnrichBidsIncludesTeamsWithZeroBidPoints(t *testing.T) {
	// GIVEN two teams where only one has a bid
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 16, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 1000},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0, "team-b": 1.0}
	totalEP := 321.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN team-b appears in the results with zero bid points
	bid := findBidByTeamID(result, "team-b")
	if bid == nil {
		t.Fatal("expected bid for team-b even though it had no bid")
	}
	if bid.BidPoints != 0 {
		t.Errorf("expected 0 bid points for team with no bid, got %d", bid.BidPoints)
	}
}

func TestThatEnrichBidsReturnsOneEntryPerTeam(t *testing.T) {
	// GIVEN three teams with bids for only two
	raw := &models.LabEntryRaw{
		Teams: threeTeamMap(),
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 500},
			{TeamID: "team-b", BidPoints: 500},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0, "team-b": 240.0, "team-c": 160.0}
	totalEP := 720.0
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN there is one enriched bid per team
	if len(result) != 3 {
		t.Errorf("expected 3 enriched bids, got %d", len(result))
	}
}

func TestThatEnrichBidsPreservesSchoolName(t *testing.T) {
	// GIVEN a team named "Duke"
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0
	totalBudget := 100

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN the school name is preserved
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.SchoolName != "Duke" {
		t.Errorf("expected school name 'Duke', got '%s'", bid.SchoolName)
	}
}

func TestThatEnrichBidsPreservesSeed(t *testing.T) {
	// GIVEN a team with seed 5
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Gonzaga", Seed: 5, Region: "West"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100},
		},
	}
	teamEP := map[string]float64{"team-a": 80.0}
	totalEP := 80.0
	totalBudget := 100

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN the seed is preserved
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.Seed != 5 {
		t.Errorf("expected seed 5, got %d", bid.Seed)
	}
}

func TestThatEnrichBidsPreservesRegion(t *testing.T) {
	// GIVEN a team in the "West" region
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Gonzaga", Seed: 5, Region: "West"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100},
		},
	}
	teamEP := map[string]float64{"team-a": 80.0}
	totalEP := 80.0
	totalBudget := 100

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN the region is preserved
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.Region != "West" {
		t.Errorf("expected region 'West', got '%s'", bid.Region)
	}
}

func TestThatEnrichBidsPreservesExpectedROIPointer(t *testing.T) {
	// GIVEN a bid with an expected ROI value
	roi := 2.5
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100, ExpectedROI: &roi},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0
	totalBudget := 100

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN the expected ROI pointer is preserved
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.ExpectedROI == nil || *bid.ExpectedROI != 2.5 {
		t.Errorf("expected expected ROI 2.5, got %v", bid.ExpectedROI)
	}
}

func TestThatEnrichBidsPreservesNilExpectedROI(t *testing.T) {
	// GIVEN a bid without an expected ROI
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100, ExpectedROI: nil},
		},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0
	totalBudget := 100

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN the expected ROI is nil
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.ExpectedROI != nil {
		t.Errorf("expected nil expected ROI, got %v", bid.ExpectedROI)
	}
}

func TestThatEnrichBidsEdgePercentIsZeroWhenNaivePointsIsZero(t *testing.T) {
	// GIVEN a team with zero expected points (naive points will be 0)
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 100},
		},
	}
	teamEP := map[string]float64{"team-a": 0.0}
	totalEP := 100.0 // nonzero total to avoid NaN
	totalBudget := 1000

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN edge percent is zero (avoids divide by zero)
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.EdgePercent != 0.0 {
		t.Errorf("expected edge percent 0.0 when naive is zero, got %f", bid.EdgePercent)
	}
}

func TestThatEnrichBidsHandlesZeroTotalBudget(t *testing.T) {
	// GIVEN all bids are zero (total budget is zero)
	raw := &models.LabEntryRaw{
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
		},
		Bids: []models.LabEntryBid{},
	}
	teamEP := map[string]float64{"team-a": 320.0}
	totalEP := 320.0
	totalBudget := 0

	// WHEN enriching bids
	result := enrichBids(raw, teamEP, totalEP, totalBudget)

	// THEN naive points is zero (no budget to allocate)
	bid := findBidByTeamID(result, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.NaivePoints != 0 {
		t.Errorf("expected 0 naive points with zero budget, got %d", bid.NaivePoints)
	}
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
// EnrichEntry (integration of internal functions)
// ---------------------------------------------------------------------------

func TestThatEnrichEntryReturnsValidResultForEmptyEntry(t *testing.T) {
	// GIVEN an entry with no bids, no predictions, and no teams
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
		Predictions:    []models.LabPrediction{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the result is not nil
	if result == nil {
		t.Error("expected non-nil result for empty entry")
	}
}

func TestThatEnrichEntryReturnsEmptyBidsSliceForEmptyEntry(t *testing.T) {
	// GIVEN an entry with no teams or bids
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
		Predictions:    []models.LabPrediction{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the bids slice is empty
	if len(result.Bids) != 0 {
		t.Errorf("expected 0 bids, got %d", len(result.Bids))
	}
}

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

func TestThatEnrichEntryPreservesEntryID(t *testing.T) {
	// GIVEN an entry with a specific ID
	raw := &models.LabEntryRaw{
		ID:             "entry-abc-123",
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the entry ID is preserved
	if result.ID != "entry-abc-123" {
		t.Errorf("expected ID 'entry-abc-123', got '%s'", result.ID)
	}
}

func TestThatEnrichEntryPreservesModelName(t *testing.T) {
	// GIVEN an entry with model name "gradient_boost_v1"
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		ModelName:      "gradient_boost_v1",
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the model name is preserved
	if result.ModelName != "gradient_boost_v1" {
		t.Errorf("expected model name 'gradient_boost_v1', got '%s'", result.ModelName)
	}
}

func TestThatEnrichEntryPreservesHasPredictions(t *testing.T) {
	// GIVEN an entry with HasPredictions = true
	raw := &models.LabEntryRaw{
		ID:              "entry-1",
		HasPredictions:  true,
		TotalPoolBudget: 1000,
		Teams:           twoTeamMap(),
		Bids:            []models.LabEntryBid{},
		Predictions: []models.LabPrediction{
			{TeamID: "team-a", PredictedMarketShare: 0.5, ExpectedPoints: 320.0},
			{TeamID: "team-b", PredictedMarketShare: 0.5, ExpectedPoints: 1.0},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN HasPredictions is preserved
	if result.HasPredictions != true {
		t.Error("expected HasPredictions to be true")
	}
}

func TestThatEnrichEntryCreatesBidForEveryTeam(t *testing.T) {
	// GIVEN three teams but only two have bids
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          threeTeamMap(),
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 600},
			{TeamID: "team-b", BidPoints: 400},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN there are bids for all three teams
	if len(result.Bids) != 3 {
		t.Errorf("expected 3 bids (one per team), got %d", len(result.Bids))
	}
}

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

	// THEN team-a (seed 1) naive allocation is proportional to seed 1 expected points
	// seed 1 = 320, seed 16 = 1, total = 321
	// naive for team-a = 320/321 * 1000 = 996 (int truncation)
	bid := findBidByTeamID(result.Bids, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	expectedNaive := int(SeedExpectedPoints[1] / (SeedExpectedPoints[1] + SeedExpectedPoints[16]) * 1000.0)
	if bid.NaivePoints != expectedNaive {
		t.Errorf("expected naive points %d, got %d", expectedNaive, bid.NaivePoints)
	}
}

func TestThatEnrichEntryUsesPredictionExpectedPointsForNaiveAllocation(t *testing.T) {
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

	// THEN bid naive allocation uses prediction expected points (600/(600+400) * 1000 = 600)
	bid := findBidByTeamID(result.Bids, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	expectedNaive := int(600.0 / 1000.0 * 1000.0)
	if bid.NaivePoints != expectedNaive {
		t.Errorf("expected naive points %d, got %d", expectedNaive, bid.NaivePoints)
	}
}

func TestThatEnrichEntryBidsAreDeterministicRegardlessOfMapIteration(t *testing.T) {
	// GIVEN an entry with multiple teams
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams:          threeTeamMap(),
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 500},
			{TeamID: "team-b", BidPoints: 300},
			{TeamID: "team-c", BidPoints: 200},
		},
	}

	// WHEN enriching the entry multiple times
	results := make([][]string, 10)
	for i := 0; i < 10; i++ {
		enriched := EnrichEntry(raw)
		teamIDs := make([]string, len(enriched.Bids))
		for j, b := range enriched.Bids {
			teamIDs[j] = b.TeamID
		}
		sort.Strings(teamIDs)
		results[i] = teamIDs
	}

	// THEN sorted team IDs are the same across all iterations
	for i := 1; i < 10; i++ {
		if len(results[i]) != len(results[0]) {
			t.Errorf("iteration %d had %d bids, expected %d", i, len(results[i]), len(results[0]))
			return
		}
		for j := range results[0] {
			if results[i][j] != results[0][j] {
				t.Errorf("iteration %d team IDs differ from iteration 0", i)
				return
			}
		}
	}
}

func TestThatEnrichEntryPreservesCalcuttaID(t *testing.T) {
	// GIVEN an entry with a specific calcutta ID
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		CalcuttaID:     "calcutta-xyz",
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the calcutta ID is preserved
	if result.CalcuttaID != "calcutta-xyz" {
		t.Errorf("expected calcutta ID 'calcutta-xyz', got '%s'", result.CalcuttaID)
	}
}

func TestThatEnrichEntryPreservesNEvaluations(t *testing.T) {
	// GIVEN an entry with 42 evaluations
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		NEvaluations:   42,
		HasPredictions: false,
		Teams:          map[string]models.LabTeamInfo{},
		Bids:           []models.LabEntryBid{},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN the evaluation count is preserved
	if result.NEvaluations != 42 {
		t.Errorf("expected 42 evaluations, got %d", result.NEvaluations)
	}
}

func TestThatEnrichEntryComputesBudgetFromBidSum(t *testing.T) {
	// GIVEN an entry with bids totaling 1000 and two equal teams
	raw := &models.LabEntryRaw{
		ID:             "entry-1",
		HasPredictions: false,
		Teams: map[string]models.LabTeamInfo{
			"team-a": {Name: "Duke", Seed: 1, Region: "East"},
			"team-b": {Name: "UNC", Seed: 1, Region: "West"},
		},
		Bids: []models.LabEntryBid{
			{TeamID: "team-a", BidPoints: 700},
			{TeamID: "team-b", BidPoints: 300},
		},
	}

	// WHEN enriching the entry
	result := EnrichEntry(raw)

	// THEN each team gets 50% naive allocation = 500 each (equal seeds)
	bid := findBidByTeamID(result.Bids, "team-a")
	if bid == nil {
		t.Fatal("expected bid for team-a")
	}
	if bid.NaivePoints != 500 {
		t.Errorf("expected naive points 500 for equal-seed team with 1000 budget, got %d", bid.NaivePoints)
	}
}
