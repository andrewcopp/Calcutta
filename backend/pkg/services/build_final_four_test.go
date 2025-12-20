package services

import (
	"testing"
)

// Tests for BracketBuilder.buildFinalFour function

func TestThatFinalFourSemifinalOneHasDeterministicID(t *testing.T) {
	// GIVEN a complete 68-team tournament
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN Final Four semifinal 1 has ID 'final_four-1'
	if bracket.Games["final_four-1"] == nil {
		t.Error("Expected Final Four semifinal with ID 'final_four-1'")
	}
}

func TestThatFinalFourSemifinalTwoHasDeterministicID(t *testing.T) {
	// GIVEN a complete 68-team tournament
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN Final Four semifinal 2 has ID 'final_four-2'
	if bracket.Games["final_four-2"] == nil {
		t.Error("Expected Final Four semifinal with ID 'final_four-2'")
	}
}

func TestThatChampionshipGameHasDeterministicID(t *testing.T) {
	// GIVEN a complete 68-team tournament
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN Championship game has ID 'championship'
	if bracket.Games["championship"] == nil {
		t.Error("Expected Championship game with ID 'championship'")
	}
}

func TestThatEastRegionalChampionLinksToFinalFourSemifinalOne(t *testing.T) {
	// GIVEN a complete 68-team tournament with East in top-left
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN East regional champion links to semifinal 1
	eastChampion := bracket.Games["East-elite_8-1"]
	if eastChampion == nil {
		t.Fatal("East regional final not found")
	}
	if eastChampion.NextGameID != "final_four-1" {
		t.Errorf("Expected East champion to link to final_four-1, got %s", eastChampion.NextGameID)
	}
}

func TestThatSouthRegionalChampionLinksToFinalFourSemifinalTwo(t *testing.T) {
	// GIVEN a complete 68-team tournament with South in top-right
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN South regional champion links to semifinal 2
	southChampion := bracket.Games["South-elite_8-1"]
	if southChampion == nil {
		t.Fatal("South regional final not found")
	}
	if southChampion.NextGameID != "final_four-2" {
		t.Errorf("Expected South champion to link to final_four-2, got %s", southChampion.NextGameID)
	}
}

func TestThatWestRegionalChampionLinksToFinalFourSemifinalOne(t *testing.T) {
	// GIVEN a complete 68-team tournament with West in bottom-left
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN West regional champion links to semifinal 1
	westChampion := bracket.Games["West-elite_8-1"]
	if westChampion == nil {
		t.Fatal("West regional final not found")
	}
	if westChampion.NextGameID != "final_four-1" {
		t.Errorf("Expected West champion to link to final_four-1, got %s", westChampion.NextGameID)
	}
}

func TestThatMidwestRegionalChampionLinksToFinalFourSemifinalTwo(t *testing.T) {
	// GIVEN a complete 68-team tournament with Midwest in bottom-right
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN Midwest regional champion links to semifinal 2
	midwestChampion := bracket.Games["Midwest-elite_8-1"]
	if midwestChampion == nil {
		t.Fatal("Midwest regional final not found")
	}
	if midwestChampion.NextGameID != "final_four-2" {
		t.Errorf("Expected Midwest champion to link to final_four-2, got %s", midwestChampion.NextGameID)
	}
}

func TestThatBothFinalFourSemifinalsLinkToChampionshipGame(t *testing.T) {
	// GIVEN a complete 68-team tournament
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()
	finalFour := helper.CreateFinalFourConfig()
	builder := NewBracketBuilder()

	// WHEN building the bracket
	bracket, err := builder.BuildBracket("tournament-1", teams, finalFour)
	if err != nil {
		t.Fatalf("Failed to build bracket: %v", err)
	}

	// THEN both semifinals link to championship
	semifinal1 := bracket.Games["final_four-1"]
	semifinal2 := bracket.Games["final_four-2"]
	if semifinal1 == nil || semifinal2 == nil {
		t.Fatal("Final Four semifinals not found")
	}
	if semifinal1.NextGameID != "championship" || semifinal2.NextGameID != "championship" {
		t.Errorf("Expected both semifinals to link to championship, got %s and %s",
			semifinal1.NextGameID, semifinal2.NextGameID)
	}
}
