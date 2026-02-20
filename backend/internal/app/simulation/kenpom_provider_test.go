package simulation

import (
	"math"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

func TestThatKenPomProviderReturnsOverrideProbabilityWhenMatchupKeyExists(t *testing.T) {
	// GIVEN a provider with an override for a specific matchup
	provider := kenPomProvider{
		overrides: map[MatchupKey]float64{
			{GameID: "g1", Team1ID: "a", Team2ID: "b"}: 0.75,
		},
	}

	// WHEN calling Prob with that matchup
	result := provider.Prob("g1", "a", "b")

	// THEN the override probability is returned
	if result != 0.75 {
		t.Errorf("expected 0.75, got %v", result)
	}
}

func TestThatKenPomProviderReturnsFiftyWhenSpecIsNil(t *testing.T) {
	// GIVEN a provider with nil spec and no overrides
	provider := kenPomProvider{
		netByTeamID: map[string]float64{"a": 10.0, "b": 5.0},
	}

	// WHEN calling Prob
	result := provider.Prob("g1", "a", "b")

	// THEN 0.5 is returned
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestThatKenPomProviderReturnsFiftyWhenTeam1RatingMissing(t *testing.T) {
	// GIVEN a provider where team1 has no rating
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	provider := kenPomProvider{
		spec:        spec,
		netByTeamID: map[string]float64{"b": 5.0},
	}

	// WHEN calling Prob with missing team1
	result := provider.Prob("g1", "a", "b")

	// THEN 0.5 is returned
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestThatKenPomProviderReturnsFiftyWhenTeam2RatingMissing(t *testing.T) {
	// GIVEN a provider where team2 has no rating
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	provider := kenPomProvider{
		spec:        spec,
		netByTeamID: map[string]float64{"a": 10.0},
	}

	// WHEN calling Prob with missing team2
	result := provider.Prob("g1", "a", "b")

	// THEN 0.5 is returned
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestThatKenPomProviderUsesSpecWinProbWhenBothRatingsPresent(t *testing.T) {
	// GIVEN a provider with both ratings and a spec (sigma=10)
	spec := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	provider := kenPomProvider{
		spec:        spec,
		netByTeamID: map[string]float64{"a": 20.0, "b": 10.0},
	}

	// WHEN calling Prob
	result := provider.Prob("g1", "a", "b")

	// THEN result equals Sigmoid((20-10)/10) = Sigmoid(1.0), which is > 0.5
	expected := 1.0 / (1.0 + math.Exp(-1.0))
	if math.Abs(result-expected) > 1e-9 {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
