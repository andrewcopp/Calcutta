package tournament

import (
	"fmt"
	"testing"
)

func TestThatBuildTeamsFromInputsAssignsUniqueIDs(t *testing.T) {
	// GIVEN two team inputs
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 1, Region: "East"},
		{SchoolID: "s2", Seed: 2, Region: "East"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN each team has a unique ID
	if teams[0].ID == teams[1].ID {
		t.Errorf("expected unique IDs, got %s and %s", teams[0].ID, teams[1].ID)
	}
}

func TestThatBuildTeamsFromInputsSetsWinsToZero(t *testing.T) {
	// GIVEN a team input
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 1, Region: "East"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN wins is 0
	if teams[0].Wins != 0 {
		t.Errorf("expected wins=0, got %d", teams[0].Wins)
	}
}

func TestThatBuildTeamsFromInputsSetsEliminatedToFalse(t *testing.T) {
	// GIVEN a team input
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 1, Region: "East"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN eliminated is false
	if teams[0].Eliminated {
		t.Errorf("expected eliminated=false")
	}
}

func TestThatBuildTeamsFromInputsSetsCorrectTournamentID(t *testing.T) {
	// GIVEN a team input
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 1, Region: "East"},
	}

	// WHEN building teams with tournament ID "t99"
	teams := buildTeamsFromInputs("t99", inputs)

	// THEN tournament ID is set correctly
	if teams[0].TournamentID != "t99" {
		t.Errorf("expected tournamentID=t99, got %s", teams[0].TournamentID)
	}
}

func TestThatBuildTeamsFromInputsSetsByesToOneForSingleTeamAtSeed(t *testing.T) {
	// GIVEN one team at seed 1 in East
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 1, Region: "East"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN byes is 1 (non-play-in)
	if teams[0].Byes != 1 {
		t.Errorf("expected byes=1, got %d", teams[0].Byes)
	}
}

func TestThatBuildTeamsFromInputsSetsByesToZeroForPlayInPair(t *testing.T) {
	// GIVEN two teams at the same seed+region (play-in)
	inputs := []ReplaceTeamsInput{
		{SchoolID: "s1", Seed: 16, Region: "East"},
		{SchoolID: "s2", Seed: 16, Region: "East"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN both teams have byes=0
	if teams[0].Byes != 0 || teams[1].Byes != 0 {
		t.Errorf("expected byes=0 for both play-in teams, got %d and %d", teams[0].Byes, teams[1].Byes)
	}
}

func TestThatBuildTeamsFromInputsPreservesSchoolAndSeedAndRegion(t *testing.T) {
	// GIVEN a team input with specific values
	inputs := []ReplaceTeamsInput{
		{SchoolID: "school-abc", Seed: 12, Region: "Midwest"},
	}

	// WHEN building teams
	teams := buildTeamsFromInputs("t1", inputs)

	// THEN school, seed, and region are preserved
	got := fmt.Sprintf("%s-%d-%s", teams[0].SchoolID, teams[0].Seed, teams[0].Region)
	want := "school-abc-12-Midwest"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}
