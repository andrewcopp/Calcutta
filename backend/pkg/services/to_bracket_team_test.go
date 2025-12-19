package services

import (
	"testing"
)

// Tests for BracketBuilder.toBracketTeam function

func TestThatBracketTeamInitiallyHasLowestSeedSeenEqualToOwnSeed(t *testing.T) {
	// GIVEN a team with seed 5
	helper := NewBracketTestHelper()
	team := helper.CreateTeam("team5", "East", 5)
	builder := NewBracketBuilder()

	// WHEN converting to bracket team
	bracketTeam := builder.toBracketTeam(team)

	// THEN LowestSeedSeen equals the team's seed
	if bracketTeam.LowestSeedSeen != 5 {
		t.Errorf("Expected LowestSeedSeen = 5, got %d", bracketTeam.LowestSeedSeen)
	}
}
