package services

import (
	"testing"
)

// Tests for BracketBuilder.BuildBracket function

func TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds(t *testing.T) {
	// GIVEN a tournament with 68 teams properly seeded
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket twice
	bracket1, err1 := builder.BuildBracket("tournament-1", teams, finalFour)
	if err1 != nil {
		t.Fatalf("First build failed: %v", err1)
	}
	bracket2, err2 := builder.BuildBracket("tournament-1", teams, finalFour)
	if err2 != nil {
		t.Fatalf("Second build failed: %v", err2)
	}

	// THEN all game IDs from first build exist in second build
	for gameID := range bracket1.Games {
		if _, exists := bracket2.Games[gameID]; !exists {
			t.Errorf("Game ID %s exists in bracket1 but not bracket2", gameID)
			return
		}
	}
}
