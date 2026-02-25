package prediction

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
)

func TestThatLatestCheckpointReturnsNilForEmptySlice(t *testing.T) {
	// GIVEN no checkpoints
	checkpoints := []CheckpointData{}

	// WHEN finding the latest checkpoint
	result := LatestCheckpoint(checkpoints)

	// THEN result is nil
	if result != nil {
		t.Errorf("expected nil, got checkpoint with ThroughRound=%d", result.ThroughRound)
	}
}

func TestThatLatestCheckpointReturnsHighestThroughRound(t *testing.T) {
	// GIVEN checkpoints at rounds 0, 2, and 5
	checkpoints := []CheckpointData{
		{ThroughRound: 0},
		{ThroughRound: 5},
		{ThroughRound: 2},
	}

	// WHEN finding the latest checkpoint
	result := LatestCheckpoint(checkpoints)

	// THEN result has ThroughRound=5
	if result.ThroughRound != 5 {
		t.Errorf("expected ThroughRound=5, got %d", result.ThroughRound)
	}
}

func TestThatBestCheckpointForCapReturnsNilWhenNoCheckpointQualifies(t *testing.T) {
	// GIVEN checkpoints all above the cap
	checkpoints := []CheckpointData{
		{ThroughRound: 3},
		{ThroughRound: 5},
	}

	// WHEN finding the best checkpoint for cap=2
	result := BestCheckpointForCap(checkpoints, 2)

	// THEN result is nil
	if result != nil {
		t.Errorf("expected nil, got checkpoint with ThroughRound=%d", result.ThroughRound)
	}
}

func TestThatBestCheckpointForCapReturnsHighestQualifyingCheckpoint(t *testing.T) {
	// GIVEN checkpoints at 0, 2, 3, and 5
	checkpoints := []CheckpointData{
		{ThroughRound: 0},
		{ThroughRound: 2},
		{ThroughRound: 3},
		{ThroughRound: 5},
	}

	// WHEN finding the best checkpoint for cap=4
	result := BestCheckpointForCap(checkpoints, 4)

	// THEN result has ThroughRound=3 (highest <= 4)
	if result.ThroughRound != 3 {
		t.Errorf("expected ThroughRound=3, got %d", result.ThroughRound)
	}
}

func TestThatBestCheckpointForCapReturnsExactMatch(t *testing.T) {
	// GIVEN checkpoints including one at the exact cap value
	checkpoints := []CheckpointData{
		{ThroughRound: 0},
		{ThroughRound: 3},
		{ThroughRound: 5},
	}

	// WHEN finding the best checkpoint for cap=3
	result := BestCheckpointForCap(checkpoints, 3)

	// THEN result has ThroughRound=3
	if result.ThroughRound != 3 {
		t.Errorf("expected ThroughRound=3, got %d", result.ThroughRound)
	}
}

func TestThatComputeEntryProjectionsReturnsNilWithNoCheckpoints(t *testing.T) {
	// GIVEN no checkpoints
	checkpoints := []CheckpointData{}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN computing entry projections
	result := ComputeEntryProjections(checkpoints, rules, nil, nil, nil)

	// THEN result is nil
	if result != nil {
		t.Error("expected nil, got non-nil projections")
	}
}

func TestThatComputeEntryProjectionsAggregatesAcrossPortfolioTeams(t *testing.T) {
	// GIVEN a checkpoint with two teams, one entry owning both through a portfolio
	checkpoints := []CheckpointData{
		{
			ThroughRound: 0,
			PTVByTeam: map[string]PredictedTeamValue{
				"team-a": {TeamID: "team-a", ExpectedPoints: 100.0, FavoritesTotalPoints: 50.0},
				"team-b": {TeamID: "team-b", ExpectedPoints: 60.0, FavoritesTotalPoints: 30.0},
			},
		},
	}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}
	portfolioToEntry := map[string]string{"port-1": "entry-1"}
	portfolioTeams := []PortfolioTeamInput{
		{PortfolioID: "port-1", TeamID: "team-a", OwnershipPercentage: 0.5},
		{PortfolioID: "port-1", TeamID: "team-b", OwnershipPercentage: 1.0},
	}
	tournamentTeams := []TournamentTeamInput{
		{ID: "team-a", Wins: 0, Byes: 0},
		{ID: "team-b", Wins: 0, Byes: 0},
	}

	// WHEN computing entry projections
	result := ComputeEntryProjections(checkpoints, rules, portfolioToEntry, portfolioTeams, tournamentTeams)

	// THEN EV = 0.5*100 + 1.0*60 = 110, Favorites = 0.5*50 + 1.0*30 = 55
	if math.Abs(result.EV["entry-1"]-110.0) > 0.001 {
		t.Errorf("expected EV=110.0, got %.2f", result.EV["entry-1"])
	}
}

func TestThatComputeEntryProjectionsAggregatesFavoritesAcrossPortfolioTeams(t *testing.T) {
	// GIVEN a checkpoint with two teams, one entry owning both
	checkpoints := []CheckpointData{
		{
			ThroughRound: 0,
			PTVByTeam: map[string]PredictedTeamValue{
				"team-a": {TeamID: "team-a", ExpectedPoints: 100.0, FavoritesTotalPoints: 50.0},
				"team-b": {TeamID: "team-b", ExpectedPoints: 60.0, FavoritesTotalPoints: 30.0},
			},
		},
	}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}
	portfolioToEntry := map[string]string{"port-1": "entry-1"}
	portfolioTeams := []PortfolioTeamInput{
		{PortfolioID: "port-1", TeamID: "team-a", OwnershipPercentage: 0.5},
		{PortfolioID: "port-1", TeamID: "team-b", OwnershipPercentage: 1.0},
	}
	tournamentTeams := []TournamentTeamInput{
		{ID: "team-a", Wins: 0, Byes: 0},
		{ID: "team-b", Wins: 0, Byes: 0},
	}

	// WHEN computing entry projections
	result := ComputeEntryProjections(checkpoints, rules, portfolioToEntry, portfolioTeams, tournamentTeams)

	// THEN Favorites = 0.5*50 + 1.0*30 = 55
	if math.Abs(result.Favorites["entry-1"]-55.0) > 0.001 {
		t.Errorf("expected Favorites=55.0, got %.2f", result.Favorites["entry-1"])
	}
}

func TestThatCapTournamentTeamsCapsWinsAtRoundCap(t *testing.T) {
	// GIVEN a team with 4 wins and round cap 2
	teams := []TournamentTeamInput{
		{ID: "t1", Wins: 4, Byes: 0, IsEliminated: false},
	}

	// WHEN capping at round 2
	result := capTournamentTeams(teams, 2)

	// THEN wins are capped at 2
	if result[0].Wins != 2 {
		t.Errorf("expected Wins=2, got %d", result[0].Wins)
	}
}

func TestThatCapTournamentTeamsZeroesByesAndFoldsIntoWins(t *testing.T) {
	// GIVEN a team with 1 win and 1 bye, round cap 3
	teams := []TournamentTeamInput{
		{ID: "t1", Wins: 1, Byes: 1, IsEliminated: false},
	}

	// WHEN capping at round 3
	result := capTournamentTeams(teams, 3)

	// THEN byes are zeroed and progress (2) is in wins
	if result[0].Wins != 2 {
		t.Errorf("expected Wins=2, got %d", result[0].Wins)
	}
	if result[0].Byes != 0 {
		t.Errorf("expected Byes=0, got %d", result[0].Byes)
	}
}

func TestThatCapTournamentTeamsTreatsEliminatedBeyondCapAsAlive(t *testing.T) {
	// GIVEN a team eliminated at progress 4, round cap 2
	teams := []TournamentTeamInput{
		{ID: "t1", Wins: 4, Byes: 0, IsEliminated: true},
	}

	// WHEN capping at round 2
	result := capTournamentTeams(teams, 2)

	// THEN team is treated as alive (progress 4 > cap 2)
	if result[0].IsEliminated {
		t.Error("expected IsEliminated=false for team eliminated beyond cap")
	}
}

func TestThatCapTournamentTeamsPreservesEliminationAtOrBelowCap(t *testing.T) {
	// GIVEN a team eliminated at progress 2, round cap 3
	teams := []TournamentTeamInput{
		{ID: "t1", Wins: 2, Byes: 0, IsEliminated: true},
	}

	// WHEN capping at round 3
	result := capTournamentTeams(teams, 3)

	// THEN elimination is preserved (progress 2 <= cap 3)
	if !result[0].IsEliminated {
		t.Error("expected IsEliminated=true for team eliminated at or below cap")
	}
}

func TestThatComputeRoundProjectionsReturnsNilWhenNoCheckpointMatchesCap(t *testing.T) {
	// GIVEN checkpoints only at round 3
	checkpoints := []CheckpointData{
		{ThroughRound: 3},
	}
	rules := []scoring.Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN computing round projections for cap=1 (no checkpoint <= 1)
	result := ComputeRoundProjections(checkpoints, rules, nil, nil, nil, 1)

	// THEN result is nil
	if result != nil {
		t.Error("expected nil, got non-nil projections")
	}
}

func TestThatComputeRoundProjectionsCapsTeamProgressAtRoundCap(t *testing.T) {
	// GIVEN a team with 3 wins but round cap is 2
	checkpoints := []CheckpointData{
		{
			ThroughRound: 0,
			PTVByTeam: map[string]PredictedTeamValue{
				"team-a": {
					TeamID:               "team-a",
					ExpectedPoints:       100.0,
					PRound1:              1.0,
					PRound2:              0.8,
					PRound3:              0.5,
					FavoritesTotalPoints: 40.0,
				},
			},
		},
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}
	portfolioToEntry := map[string]string{"port-1": "entry-1"}
	portfolioTeams := []PortfolioTeamInput{
		{PortfolioID: "port-1", TeamID: "team-a", OwnershipPercentage: 1.0},
	}
	tournamentTeams := []TournamentTeamInput{
		{ID: "team-a", Wins: 3, Byes: 0, IsEliminated: false},
	}

	// WHEN computing projections at round cap 2
	result := ComputeRoundProjections(checkpoints, rules, portfolioToEntry, portfolioTeams, tournamentTeams, 2)

	// THEN team progress is capped at 2 wins, so actual = 10+20 = 30
	// The EV should reflect capped progress, not full 3-win progress
	if result == nil {
		t.Fatal("expected non-nil projections")
	}
	// EV should be based on capped progress (2 wins), not 3
	// With pAlive=PRound2=0.8, conditional remaining round 3: (0.5/0.8)*40 = 25.0
	// total = 30 + 25 = 55.0
	expected := 55.0
	if math.Abs(result.EV["entry-1"]-expected) > 0.001 {
		t.Errorf("expected EV=%.2f, got %.2f", expected, result.EV["entry-1"])
	}
}

func TestThatComputeRoundProjectionsTreatsEliminatedTeamBeyondCapAsAlive(t *testing.T) {
	// GIVEN a team eliminated at progress 4, but round cap is 2
	checkpoints := []CheckpointData{
		{
			ThroughRound: 0,
			PTVByTeam: map[string]PredictedTeamValue{
				"team-a": {
					TeamID:               "team-a",
					ExpectedPoints:       100.0,
					PRound1:              1.0,
					PRound2:              0.8,
					PRound3:              0.5,
					FavoritesTotalPoints: 40.0,
				},
			},
		},
	}
	rules := []scoring.Rule{
		{WinIndex: 1, PointsAwarded: 10},
		{WinIndex: 2, PointsAwarded: 20},
		{WinIndex: 3, PointsAwarded: 40},
	}
	portfolioToEntry := map[string]string{"port-1": "entry-1"}
	portfolioTeams := []PortfolioTeamInput{
		{PortfolioID: "port-1", TeamID: "team-a", OwnershipPercentage: 1.0},
	}
	tournamentTeams := []TournamentTeamInput{
		// Eliminated at progress 4, but cap is 2 → isEliminated should be false
		{ID: "team-a", Wins: 4, Byes: 0, IsEliminated: true},
	}

	// WHEN computing projections at round cap 2
	result := ComputeRoundProjections(checkpoints, rules, portfolioToEntry, portfolioTeams, tournamentTeams, 2)

	// THEN team is NOT considered eliminated (progress 4 > cap 2)
	// So it gets projected EV, not just actual points
	if result == nil {
		t.Fatal("expected non-nil projections")
	}
	// Capped progress = 2, pAlive = PRound2 = 0.8
	// conditional remaining round 3: (0.5/0.8)*40 = 25.0
	// total = 30 + 25 = 55.0
	expected := 55.0
	if math.Abs(result.EV["entry-1"]-expected) > 0.001 {
		t.Errorf("expected EV=%.2f, got %.2f", expected, result.EV["entry-1"])
	}
}
