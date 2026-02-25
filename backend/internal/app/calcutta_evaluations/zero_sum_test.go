package calcutta_evaluations

import (
	"math"
	"testing"
)

// addHouseEntry creates a "house" entry bidding 1 point on every unclaimed team.
// This absorbs points from teams that no real entry bid on, closing the leak.
func addHouseEntry(entries map[string]*Entry, allTeamIDs []string) map[string]*Entry {
	claimed := make(map[string]bool)
	for _, entry := range entries {
		for teamID := range entry.Teams {
			claimed[teamID] = true
		}
	}

	houseTeams := make(map[string]int)
	for _, id := range allTeamIDs {
		if !claimed[id] {
			houseTeams[id] = 1
		}
	}

	if len(houseTeams) == 0 {
		return entries
	}

	result := make(map[string]*Entry, len(entries)+1)
	for k, v := range entries {
		result[k] = v
	}
	result["house"] = &Entry{Name: "House", Teams: houseTeams}
	return result
}

func sumEntryPoints(results []SimulationResult) float64 {
	var total float64
	for _, r := range results {
		total += r.TotalPoints
	}
	return total
}

func sumTeamPoints(teamResults []TeamSimResult) float64 {
	var total float64
	for _, tr := range teamResults {
		total += float64(tr.Points)
	}
	return total
}

func TestThatEntryPointsEqualTeamPointsWhenAllTeamsClaimed(t *testing.T) {
	// GIVEN entries that collectively bid on every team
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 60, "teamC": 30}},
		"bob":   {Name: "Bob", Teams: map[string]int{"teamA": 40, "teamB": 100}},
		"carol": {Name: "Carol", Teams: map[string]int{"teamC": 70, "teamD": 100}},
	}
	teamResults := []TeamSimResult{
		{TeamID: "teamA", Points: 100},
		{TeamID: "teamB", Points: 50},
		{TeamID: "teamC", Points: 200},
		{TeamID: "teamD", Points: 30},
	}

	// WHEN calculating simulation outcomes
	results, err := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN sum of entry points equals sum of team points
	entrySum := sumEntryPoints(results)
	teamSum := sumTeamPoints(teamResults)
	if math.Abs(entrySum-teamSum) > 1e-9 {
		t.Errorf("entry sum = %.4f, team sum = %.4f", entrySum, teamSum)
	}
}

func TestThatHouseEntryAbsorbsUnclaimedTeamPoints(t *testing.T) {
	// GIVEN entries that do NOT bid on teamD
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
		"bob":   {Name: "Bob", Teams: map[string]int{"teamB": 100}},
	}
	allTeamIDs := []string{"teamA", "teamB", "teamC", "teamD"}
	teamResults := []TeamSimResult{
		{TeamID: "teamA", Points: 100},
		{TeamID: "teamB", Points: 50},
		{TeamID: "teamC", Points: 200},
		{TeamID: "teamD", Points: 30},
	}

	// WHEN adding a house entry and calculating outcomes
	withHouse := addHouseEntry(entries, allTeamIDs)
	results, err := CalculateSimulationOutcomes(1, withHouse, teamResults, map[int]int{}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN entry sum (including house) equals team sum
	entrySum := sumEntryPoints(results)
	teamSum := sumTeamPoints(teamResults)
	if math.Abs(entrySum-teamSum) > 1e-9 {
		t.Errorf("entry sum = %.4f, team sum = %.4f", entrySum, teamSum)
	}
}

func TestThatUnclaimedTeamPointsLeakWithoutHouse(t *testing.T) {
	// GIVEN entries that do NOT bid on teamC or teamD
	entries := map[string]*Entry{
		"alice": {Name: "Alice", Teams: map[string]int{"teamA": 100}},
		"bob":   {Name: "Bob", Teams: map[string]int{"teamB": 100}},
	}
	teamResults := []TeamSimResult{
		{TeamID: "teamA", Points: 100},
		{TeamID: "teamB", Points: 50},
		{TeamID: "teamC", Points: 200},
		{TeamID: "teamD", Points: 30},
	}

	// WHEN calculating without a house entry
	results, err := CalculateSimulationOutcomes(1, entries, teamResults, map[int]int{}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN entry sum is LESS than team sum (points leaked)
	entrySum := sumEntryPoints(results)
	teamSum := sumTeamPoints(teamResults)
	if entrySum >= teamSum {
		t.Errorf("expected entry sum (%.4f) < team sum (%.4f) when teams are unclaimed",
			entrySum, teamSum)
	}
}
