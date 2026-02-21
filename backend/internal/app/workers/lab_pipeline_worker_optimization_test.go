package workers

import (
	"encoding/json"
	"testing"
)

func TestThatValidAllocationReturnsNoError(t *testing.T) {
	// GIVEN a valid allocation within all constraints
	bids := map[string]int{"team1": 30, "team2": 20, "team3": 10}
	constraints := optimizationConstraints{MinTeams: 2, MaxTeams: 5, MaxPerTeam: 50, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatAllocationExceedingBudgetReturnsError(t *testing.T) {
	// GIVEN bids that exceed the budget
	bids := map[string]int{"team1": 60, "team2": 50}
	constraints := optimizationConstraints{MinTeams: 1, MaxTeams: 5, MaxPerTeam: 100, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN an error about budget violation is returned
	if err == nil {
		t.Error("expected budget violation error, got nil")
	}
}

func TestThatAllocationBelowMinTeamsReturnsError(t *testing.T) {
	// GIVEN bids with fewer teams than the minimum
	bids := map[string]int{"team1": 50}
	constraints := optimizationConstraints{MinTeams: 3, MaxTeams: 5, MaxPerTeam: 100, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN an error about min_teams violation is returned
	if err == nil {
		t.Error("expected min_teams violation error, got nil")
	}
}

func TestThatAllocationExceedingMaxTeamsReturnsError(t *testing.T) {
	// GIVEN bids with more teams than the maximum
	bids := map[string]int{"t1": 10, "t2": 10, "t3": 10, "t4": 10, "t5": 10, "t6": 10}
	constraints := optimizationConstraints{MinTeams: 1, MaxTeams: 3, MaxPerTeam: 100, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN an error about max_teams violation is returned
	if err == nil {
		t.Error("expected max_teams violation error, got nil")
	}
}

func TestThatAllocationExceedingPerTeamMaxReturnsError(t *testing.T) {
	// GIVEN a bid that exceeds the per-team maximum
	bids := map[string]int{"team1": 60, "team2": 10}
	constraints := optimizationConstraints{MinTeams: 1, MaxTeams: 5, MaxPerTeam: 50, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN an error about max_per_team violation is returned
	if err == nil {
		t.Error("expected max_per_team violation error, got nil")
	}
}

func TestThatEmptyAllocationPassesValidation(t *testing.T) {
	// GIVEN an empty bid set (no teams selected)
	bids := map[string]int{}
	constraints := optimizationConstraints{MinTeams: 3, MaxTeams: 5, MaxPerTeam: 50, TotalPoolBudget: 500}

	// WHEN validating the allocation
	err := validateAllocation(bids, 100, constraints)

	// THEN no error is returned (min_teams only applies when numTeams > 0)
	if err != nil {
		t.Errorf("expected no error for empty bids, got %v", err)
	}
}

func TestThatBuildBidsJSONIncludesOnlyNonZeroBids(t *testing.T) {
	// GIVEN predictions for three teams but bids for only two
	predictions := []optimizationPrediction{
		{TeamID: "t1", PredictedMarketShare: 0.3, ExpectedPoints: 10.0},
		{TeamID: "t2", PredictedMarketShare: 0.2, ExpectedPoints: 5.0},
		{TeamID: "t3", PredictedMarketShare: 0.5, ExpectedPoints: 15.0},
	}
	bids := map[string]int{"t1": 30, "t3": 50}

	// WHEN building the bids JSON
	data, err := buildBidsJSON(predictions, bids, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN only teams with non-zero bids are included
	var rows []optimizationBidRow
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 bid rows, got %d", len(rows))
	}
}

func TestThatBuildBidsJSONPreservesTeamOrder(t *testing.T) {
	// GIVEN predictions in a specific order
	predictions := []optimizationPrediction{
		{TeamID: "t1", PredictedMarketShare: 0.3, ExpectedPoints: 10.0},
		{TeamID: "t2", PredictedMarketShare: 0.2, ExpectedPoints: 5.0},
	}
	bids := map[string]int{"t1": 30, "t2": 20}

	// WHEN building the bids JSON
	data, err := buildBidsJSON(predictions, bids, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the order follows the predictions input order
	var rows []optimizationBidRow
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if rows[0].TeamID != "t1" {
		t.Errorf("expected first team to be t1, got %s", rows[0].TeamID)
	}
}

func TestThatBuildBidsJSONCalculatesExpectedROI(t *testing.T) {
	// GIVEN a single prediction and bid
	predictions := []optimizationPrediction{
		{TeamID: "t1", PredictedMarketShare: 0.5, ExpectedPoints: 10.0},
	}
	bids := map[string]int{"t1": 25}

	// WHEN building the bids JSON
	data, err := buildBidsJSON(predictions, bids, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN expectedROI = expectedPoints / (marketCost + bid) = 10 / (50 + 25) = 0.1333...
	var rows []optimizationBidRow
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	expected := 10.0 / (0.5*100.0 + 25.0)
	diff := rows[0].ExpectedROI - expected
	if diff < -0.0001 || diff > 0.0001 {
		t.Errorf("expected ROI ~%.4f, got %.4f", expected, rows[0].ExpectedROI)
	}
}
