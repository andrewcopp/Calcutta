package services

import (
	"testing"
)

// Tests for BracketBuilder.buildRegionalBracket - Regional rounds (Round of 32, Sweet 16, Elite 8)

func TestThatRoundOfThirtyTwoGameHasDeterministicIDBasedOnLowestSeed(t *testing.T) {
	// GIVEN a complete regional bracket
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("South", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "South", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN Round of 32 game for 1-seed path has ID 'South-round_of_32-1'
	if bracket.Games["South-round_of_32-1"] == nil {
		t.Error("Expected Round of 32 game with ID 'South-round_of_32-1'")
	}
}

func TestThatSweetSixteenGameHasDeterministicIDBasedOnLowestSeed(t *testing.T) {
	// GIVEN a complete regional bracket
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("South", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "South", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN Sweet 16 game for 1-seed path has ID 'South-sweet_16-1'
	if bracket.Games["South-sweet_16-1"] == nil {
		t.Error("Expected Sweet 16 game with ID 'South-sweet_16-1'")
	}
}

func TestThatEliteEightGameHasDeterministicIDForRegionalFinal(t *testing.T) {
	// GIVEN a complete regional bracket
	helper := NewBracketTestHelper()
	teams := helper.CreateStandardRegionTeams("East", 16)
	builder := NewBracketBuilder()
	bracket := helper.CreateEmptyBracket()

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN Elite 8 (regional final) has ID 'East-elite_8-1'
	if bracket.Games["East-elite_8-1"] == nil {
		t.Error("Expected Elite 8 game with ID 'East-elite_8-1'")
	}
}
