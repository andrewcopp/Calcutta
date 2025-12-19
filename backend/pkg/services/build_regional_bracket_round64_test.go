package services

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// Tests for BracketBuilder.buildRegionalBracket - Round of 64 matchups

func TestThatRoundOfSixtyFourMatchesOneSeedAgainstSixteenSeed(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("West", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN the 1 vs 16 game exists with correct seeds
	game := bracket.Games["West-round_of_64-1"]
	if game == nil {
		t.Fatal("Expected game West-round_of_64-1 to exist")
	}

	seed1 := game.Team1.Seed
	seed2 := game.Team2.Seed
	if (seed1 != 1 || seed2 != 16) && (seed1 != 16 || seed2 != 1) {
		t.Errorf("Expected 1 vs 16 matchup, got %d vs %d", seed1, seed2)
	}
}

func TestThatRoundOfSixtyFourMatchesTwoSeedAgainstFifteenSeed(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("West", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN the 2 vs 15 game exists with correct seeds
	game := bracket.Games["West-round_of_64-2"]
	if game == nil {
		t.Fatal("Expected game West-round_of_64-2 to exist")
	}

	seed1 := game.Team1.Seed
	seed2 := game.Team2.Seed
	if (seed1 != 2 || seed2 != 15) && (seed1 != 15 || seed2 != 2) {
		t.Errorf("Expected 2 vs 15 matchup, got %d vs %d", seed1, seed2)
	}
}

func TestThatRoundOfSixtyFourMatchesEightSeedAgainstNineSeed(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("West", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN the 8 vs 9 game exists with correct seeds
	game := bracket.Games["West-round_of_64-8"]
	if game == nil {
		t.Fatal("Expected game West-round_of_64-8 to exist")
	}

	seed1 := game.Team1.Seed
	seed2 := game.Team2.Seed
	if (seed1 != 8 || seed2 != 9) && (seed1 != 9 || seed2 != 8) {
		t.Errorf("Expected 8 vs 9 matchup, got %d vs %d", seed1, seed2)
	}
}

func TestThatRoundOfSixtyFourSeedsSumToSeventeen(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("West", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN all Round of 64 games have seeds that sum to 17
	for _, game := range bracket.Games {
		if game.Round == models.RoundOf64 && game.Team1 != nil && game.Team2 != nil {
			seedSum := game.Team1.Seed + game.Team2.Seed
			if seedSum != 17 {
				t.Errorf("Game %s: seeds %d + %d = %d, expected 17",
					game.GameID, game.Team1.Seed, game.Team2.Seed, seedSum)
				return
			}
		}
	}
}

func TestThatRoundOfSixtyFourCreatesExactlyEightGamesPerRegion(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("South", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "South", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be exactly 8 Round of 64 games
	round64Count := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundOf64 {
			round64Count++
		}
	}
	if round64Count != 8 {
		t.Errorf("Expected 8 Round of 64 games, got %d", round64Count)
	}
}

func TestThatRoundOfSixtyFourGameForOneSeedHasDeterministicID(t *testing.T) {
	// GIVEN a region with standard seeding
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("Midwest", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "Midwest", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN the 1 vs 16 game has ID 'Midwest-round_of_64-1'
	if bracket.Games["Midwest-round_of_64-1"] == nil {
		t.Error("Expected Round of 64 game with ID 'Midwest-round_of_64-1'")
	}
}
