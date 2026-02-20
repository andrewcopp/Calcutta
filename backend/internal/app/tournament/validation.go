package tournament

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// BracketValidationError holds multiple validation errors for bracket setup.
type BracketValidationError struct {
	Errors []string
}

func (e *BracketValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0]
	}
	return fmt.Sprintf("bracket validation failed with %d errors", len(e.Errors))
}

// ValidateBracketSetup validates a set of teams for a 68-team tournament bracket.
// Returns ALL errors found, not just the first.
func ValidateBracketSetup(teams []*models.TournamentTeam) []string {
	var errs []string

	if len(teams) != 68 {
		errs = append(errs, fmt.Sprintf("tournament must have exactly 68 teams, has %d", len(teams)))
	}

	// Check for duplicate schools
	schoolIDs := make(map[string]int)
	for _, team := range teams {
		schoolIDs[team.SchoolID]++
	}
	for schoolID, count := range schoolIDs {
		if count > 1 {
			errs = append(errs, fmt.Sprintf("school %s appears %d times", schoolID, count))
		}
	}

	// Derive regions from submitted teams
	regionSet := make(map[string]bool)
	for _, team := range teams {
		if strings.TrimSpace(team.Region) == "" {
			errs = append(errs, fmt.Sprintf("team %s has empty region", team.SchoolID))
		} else {
			regionSet[team.Region] = true
		}
	}
	expectedRegions := make([]string, 0, len(regionSet))
	for r := range regionSet {
		expectedRegions = append(expectedRegions, r)
	}
	sort.Strings(expectedRegions)
	if len(expectedRegions) != 4 {
		errs = append(errs, fmt.Sprintf("must have exactly 4 regions, found %d", len(expectedRegions)))
	}

	// Check region counts
	regionCounts := make(map[string]int)
	for _, team := range teams {
		regionCounts[team.Region]++
	}
	for _, region := range expectedRegions {
		count := regionCounts[region]
		if count < 16 {
			errs = append(errs, fmt.Sprintf("region %s must have at least 16 teams, has %d", region, count))
		}
	}

	// Check seeds per region
	type regionSeed struct {
		region string
		seed   int
	}
	regionSeedCounts := make(map[regionSeed]int)
	for _, team := range teams {
		regionSeedCounts[regionSeed{region: team.Region, seed: team.Seed}]++
	}

	for _, region := range expectedRegions {
		for seed := 1; seed <= 16; seed++ {
			key := regionSeed{region: region, seed: seed}
			count := regionSeedCounts[key]
			if count == 0 {
				errs = append(errs, fmt.Sprintf("region %s is missing seed %d", region, seed))
			}
			if count > 2 {
				errs = append(errs, fmt.Sprintf("region %s has %d teams with seed %d (max 2)", region, count, seed))
			}
		}
	}

	// Check play-in count (exactly 4 seeds should have 2 teams)
	playInCount := 0
	for _, region := range expectedRegions {
		for seed := 1; seed <= 16; seed++ {
			key := regionSeed{region: region, seed: seed}
			if regionSeedCounts[key] == 2 {
				playInCount++
			}
		}
	}
	if playInCount != 4 {
		errs = append(errs, fmt.Sprintf("must have exactly 4 play-in games, has %d", playInCount))
	}

	// Check byes correctness
	for _, team := range teams {
		key := regionSeed{region: team.Region, seed: team.Seed}
		count := regionSeedCounts[key]
		if count == 2 && team.Byes != 0 {
			errs = append(errs, fmt.Sprintf("play-in team must have byes=0 (region=%s seed=%d school=%s)", team.Region, team.Seed, team.SchoolID))
		}
		if count == 1 && team.Byes != 1 {
			errs = append(errs, fmt.Sprintf("non-play-in team must have byes=1 (region=%s seed=%d school=%s)", team.Region, team.Seed, team.SchoolID))
		}
	}

	return errs
}
