package services

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// Tests for BracketBuilder.buildRegionalBracket - First Four game creation

func TestThatRegionWithNoDuplicateSeedsCreatesZeroFirstFourGames(t *testing.T) {
	// GIVEN a region with no duplicate seeds
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("East", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be zero First Four games
	firstFourCount := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundFirstFour {
			firstFourCount++
		}
	}
	if firstFourCount != 0 {
		t.Errorf("Expected 0 First Four games, got %d", firstFourCount)
	}
}

func TestThatRegionWithOneDuplicateSeedCreatesOneFirstFourGame(t *testing.T) {
	// GIVEN a region with one duplicate seed (11)
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("West", 11)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be exactly 1 First Four game
	firstFourCount := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundFirstFour {
			firstFourCount++
		}
	}
	if firstFourCount != 1 {
		t.Errorf("Expected 1 First Four game, got %d", firstFourCount)
	}
}

func TestThatRegionWithTwoDuplicateSeedsCreatesExactlyTwoFirstFourGames(t *testing.T) {
	// GIVEN a region with two duplicate seeds (11 and 16)
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be exactly 2 First Four games
	firstFourCount := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundFirstFour {
			firstFourCount++
		}
	}
	if firstFourCount != 2 {
		t.Errorf("Expected 2 First Four games, got %d", firstFourCount)
	}
}

func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T) {
	// GIVEN a region with two 11-seeds
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN First Four game for 11-seeds has ID 'East-first_four-11'
	if bracket.Games["East-first_four-11"] == nil {
		t.Error("Expected First Four game with ID 'East-first_four-11'")
	}
}

func TestThatFirstFourGameForSixteenSeedHasDeterministicID(t *testing.T) {
	// GIVEN a region with two 16-seeds
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("South", 11, 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "South", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN First Four game for 16-seeds has ID 'South-first_four-16'
	if bracket.Games["South-first_four-16"] == nil {
		t.Error("Expected First Four game with ID 'South-first_four-16'")
	}
}

func TestThatFirstFourGameMatchesBothTeamsWithSameSeed(t *testing.T) {
	// GIVEN a region with two 11-seeds
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN both teams in the 11-seed First Four game have seed 11
	game11 := bracket.Games["East-first_four-11"]
	if game11 == nil {
		t.Fatal("First Four game not found")
	}
	if game11.Team1.Seed != 11 || game11.Team2.Seed != 11 {
		t.Errorf("Expected both teams to have seed 11, got %d vs %d", game11.Team1.Seed, game11.Team2.Seed)
	}
}

func TestThatFirstFourGameLinksToCorrectRoundOfSixtyFourGame(t *testing.T) {
	// GIVEN a region with two 11-seeds
	helper := NewBracketTestHelper()
	teams := helper.CreateRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN First Four game links to the 6 vs 11 Round of 64 game
	game11 := bracket.Games["East-first_four-11"]
	if game11 == nil {
		t.Fatal("First Four game not found")
	}
	if game11.NextGameID != "East-round_of_64-6" {
		t.Errorf("Expected First Four game to link to 'East-round_of_64-6', got '%s'", game11.NextGameID)
	}
}
