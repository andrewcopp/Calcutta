package bracket

import (
	"testing"
)

func TestThatValidateBracketSetupTeamsReturnsErrorWhenPlayInTeamHasBye(t *testing.T) {
	// GIVEN a tournament with a duplicated (region, seed) but a play-in team has byes=1
	teams := createFullTournamentTeams("t")
	for _, team := range teams {
		if team != nil && team.Region == "East" && team.Seed == 11 {
			team.Byes = 1
			break
		}
	}

	// WHEN validating bracket setup
	err := ValidateBracketSetupTeams(teams)

	// THEN validation fails
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestThatValidateBracketSetupTeamsReturnsErrorWhenNonPlayInTeamHasNoBye(t *testing.T) {
	// GIVEN a tournament with a non-play-in team but byes=0
	teams := createFullTournamentTeams("t")
	for _, team := range teams {
		if team != nil && team.Region == "West" && team.Seed == 1 {
			team.Byes = 0
			break
		}
	}

	// WHEN validating bracket setup
	err := ValidateBracketSetupTeams(teams)

	// THEN validation fails
	if err == nil {
		t.Errorf("expected error")
	}
}
