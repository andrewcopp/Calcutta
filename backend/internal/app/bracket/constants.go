package bracket

import "github.com/andrewcopp/Calcutta/backend/internal/models"

const TotalTournamentTeams = 68
const TeamsPerRegion = 16
const MaxTeamsPerSeed = 2

// Seed pair sums: matched seeds in each round sum to these values.
const (
	SeedPairSumR64 = 17 // 1v16, 2v15 ... 8v9
	SeedPairSumR32 = 9  // 1v8, 2v7, 3v6, 4v5
	SeedPairSumS16 = 5  // 1v4, 2v3
	SeedPairSumE8  = 3  // 1v2
)

var Regions = []string{"East", "West", "South", "Midwest"}

const RegionSortMultiplier = 1000

var regionSortOrder = map[string]int{
	"East": 0, "West": 1, "South": 2, "Midwest": 3,
}

var roundSortOffset = map[models.BracketRound]int{
	models.RoundFirstFour:    0,
	models.RoundOf64:         100,
	models.RoundOf32:         200,
	models.RoundSweet16:      300,
	models.RoundElite8:       400,
	models.RoundFinalFour:    500,
	models.RoundChampionship: 600,
}
