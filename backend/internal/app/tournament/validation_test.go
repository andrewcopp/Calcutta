package tournament

import (
	"fmt"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// createValidTeams creates a valid 68-team set with the standard NCAA tournament structure.
// playInSeeds maps region name to the seed that has 2 teams (play-in).
func createValidTeams(playInSeeds map[string]int) []*models.TournamentTeam {
	regions := []string{"East", "West", "South", "Midwest"}
	if playInSeeds == nil {
		playInSeeds = map[string]int{
			"East":    16,
			"West":    16,
			"South":   11,
			"Midwest": 11,
		}
	}

	teams := make([]*models.TournamentTeam, 0, 68)
	schoolIdx := 0
	for _, region := range regions {
		playInSeed := playInSeeds[region]
		for seed := 1; seed <= 16; seed++ {
			count := 1
			if seed == playInSeed {
				count = 2
			}
			for i := 0; i < count; i++ {
				byes := 1
				if count == 2 {
					byes = 0
				}
				teams = append(teams, &models.TournamentTeam{
					ID:           fmt.Sprintf("team-%d", schoolIdx),
					TournamentID: "t1",
					SchoolID:     fmt.Sprintf("school-%d", schoolIdx),
					Seed:         seed,
					Region:       region,
					Byes:         byes,
				})
				schoolIdx++
			}
		}
	}
	return teams
}

func TestThatValidBracketSetupReturnsNoErrors(t *testing.T) {
	// GIVEN a valid 68-team bracket
	teams := createValidTeams(nil)

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN no errors are returned
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForWrongTeamCount(t *testing.T) {
	// GIVEN a bracket with 67 teams
	teams := createValidTeams(nil)
	teams = teams[:67]

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about team count is returned
	if !containsError(errs, "must have exactly 68 teams") {
		t.Errorf("expected team count error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForDuplicateSchool(t *testing.T) {
	// GIVEN a valid bracket with one school duplicated
	teams := createValidTeams(nil)
	teams[0].SchoolID = teams[1].SchoolID

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about duplicate school is returned
	if !containsError(errs, "appears 2 times") {
		t.Errorf("expected duplicate school error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForMissingSeed(t *testing.T) {
	// GIVEN a bracket where East is missing seed 5 (replaced with extra seed 6)
	teams := createValidTeams(nil)
	for i, team := range teams {
		if team.Region == "East" && team.Seed == 5 {
			teams[i].Seed = 6
			break
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about missing seed is returned
	if !containsError(errs, "missing seed 5") {
		t.Errorf("expected missing seed error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForTooManyTeamsAtSeed(t *testing.T) {
	// GIVEN a valid bracket modified so East seed 3 has 3 teams
	teams := createValidTeams(nil)
	// Change two East teams (seed 1 and 2) to seed 3, giving seed 3 a total of 3 teams
	changed := 0
	for i, team := range teams {
		if team.Region == "East" && (team.Seed == 1 || team.Seed == 2) && changed < 2 {
			teams[i].Seed = 3
			changed++
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about too many teams at seed is returned
	if !containsError(errs, "has 3 teams with seed 3") {
		t.Errorf("expected too-many-at-seed error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForWrongPlayInCount(t *testing.T) {
	// GIVEN a bracket with 5 play-in games instead of 4
	playIns := map[string]int{
		"East":    16,
		"West":    16,
		"South":   11,
		"Midwest": 16, // 3 regions with play-in at 16
	}
	teams := createValidTeams(playIns)
	// Add one more play-in for South seed 16
	teams = append(teams, &models.TournamentTeam{
		ID: "extra", TournamentID: "t1", SchoolID: "extra-school",
		Seed: 16, Region: "South", Byes: 0,
	})
	// Fix byes for existing South seed 16
	for i, team := range teams {
		if team.Region == "South" && team.Seed == 16 {
			teams[i].Byes = 0
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about play-in count is returned
	if !containsError(errs, "play-in games") {
		t.Errorf("expected play-in count error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForPlayInTeamWithByes(t *testing.T) {
	// GIVEN a valid bracket where a play-in team has byes=1
	teams := createValidTeams(nil)
	for i, team := range teams {
		if team.Byes == 0 {
			teams[i].Byes = 1
			break
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about play-in team byes is returned
	if !containsError(errs, "play-in team must have byes=0") {
		t.Errorf("expected play-in byes error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsErrorForNonPlayInTeamWithoutByes(t *testing.T) {
	// GIVEN a valid bracket where a non-play-in team has byes=0
	teams := createValidTeams(nil)
	for i, team := range teams {
		if team.Byes == 1 {
			teams[i].Byes = 0
			break
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about non-play-in team byes is returned
	if !containsError(errs, "non-play-in team must have byes=1") {
		t.Errorf("expected non-play-in byes error, got %v", errs)
	}
}

func TestThatBracketSetupReturnsMultipleErrors(t *testing.T) {
	// GIVEN an empty team list
	teams := []*models.TournamentTeam{}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN multiple errors are returned (at least team count + region errors)
	if len(errs) < 2 {
		t.Errorf("expected multiple errors, got %d: %v", len(errs), errs)
	}
}

func TestThatBracketSetupReturnsErrorForEmptyRegion(t *testing.T) {
	// GIVEN a bracket with no East teams (all moved to a non-standard region)
	teams := createValidTeams(nil)
	for i, team := range teams {
		if team.Region == "East" {
			teams[i].Region = "North"
		}
	}

	// WHEN validating
	errs := ValidateBracketSetup(teams)

	// THEN an error about East region is returned
	if !containsError(errs, "East") {
		t.Errorf("expected East region error, got %v", errs)
	}
}

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if contains(e, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
