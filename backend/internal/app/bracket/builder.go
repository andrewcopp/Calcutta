package bracket

import (
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type BracketBuilder struct{}

func NewBracketBuilder() *BracketBuilder {
	return &BracketBuilder{}
}

func (b *BracketBuilder) BuildBracket(tournamentID string, teams []*models.TournamentTeam, finalFour *models.FinalFourConfig) (*models.BracketStructure, error) {
	return BuildBracketStructure(tournamentID, teams, finalFour)
}

func BuildBracketStructure(tournamentID string, teams []*models.TournamentTeam, finalFour *models.FinalFourConfig) (*models.BracketStructure, error) {
	b := &BracketBuilder{}

	if len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(teams))
	}

	bracket := &models.BracketStructure{
		TournamentID: tournamentID,
		Regions:      []string{"East", "West", "South", "Midwest"},
		Games:        make(map[string]*models.BracketGame),
		FinalFour:    finalFour,
	}

	teamsByRegion := make(map[string][]*models.TournamentTeam)
	for _, team := range teams {
		teamsByRegion[team.Region] = append(teamsByRegion[team.Region], team)
	}

	regionalChampionGameIDs := make(map[string]string)
	for _, region := range bracket.Regions {
		regionTeams := teamsByRegion[region]
		championGameID, err := b.buildRegionalBracket(bracket, region, regionTeams)
		if err != nil {
			return nil, fmt.Errorf("error building %s region: %w", region, err)
		}
		regionalChampionGameIDs[region] = championGameID
	}

	if err := b.buildFinalFour(bracket, regionalChampionGameIDs, finalFour); err != nil {
		return nil, fmt.Errorf("error building Final Four: %w", err)
	}

	return bracket, nil
}

func (b *BracketBuilder) buildRegionalBracket(bracket *models.BracketStructure, region string, teams []*models.TournamentTeam) (string, error) {
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Seed < teams[j].Seed
	})

	firstFourTeams := b.identifyFirstFourTeams(teams)

	firstFourWinners := make(map[int]string)
	for seed, seedTeams := range firstFourTeams {
		if len(seedTeams) == 2 {
			gameID := fmt.Sprintf("%s-first_four-%d", region, seed)
			game := &models.BracketGame{
				GameID:    gameID,
				Round:     models.RoundFirstFour,
				Region:    region,
				Team1:     b.toBracketTeam(seedTeams[0]),
				Team2:     b.toBracketTeam(seedTeams[1]),
				SortOrder: b.getSortOrder(models.RoundFirstFour, region, seed),
			}
			bracket.Games[gameID] = game
			firstFourWinners[seed] = gameID
		}
	}

	round64Games := b.buildRoundOf64(bracket, region, teams, firstFourWinners)
	round32Games := b.buildRegionalRound(bracket, region, models.RoundOf32, round64Games, 9)
	sweet16Games := b.buildRegionalRound(bracket, region, models.RoundSweet16, round32Games, 5)
	elite8Games := b.buildRegionalRound(bracket, region, models.RoundElite8, sweet16Games, 3)
	championGameID := b.buildRegionalChampionship(bracket, region, elite8Games)

	return championGameID, nil
}

func (b *BracketBuilder) buildRoundOf64(bracket *models.BracketStructure, region string, teams []*models.TournamentTeam, firstFourWinners map[int]string) map[int]string {
	round64Games := make(map[int]string)
	teamsWithByes := b.getTeamsWithByes(teams, b.identifyFirstFourTeams(teams))

	processed := make(map[int]bool)
	for _, team := range teamsWithByes {
		if processed[team.Seed] {
			continue
		}

		opponentSeed := 17 - team.Seed
		if processed[opponentSeed] {
			continue
		}

		lowestSeed := team.Seed
		if opponentSeed < team.Seed {
			lowestSeed = opponentSeed
		}
		gameID := fmt.Sprintf("%s-round_of_64-%d", region, lowestSeed)
		game := &models.BracketGame{
			GameID:    gameID,
			Round:     models.RoundOf64,
			Region:    region,
			SortOrder: b.getSortOrder(models.RoundOf64, region, team.Seed),
		}

		if team.Seed < opponentSeed {
			game.Team1 = b.toBracketTeam(team)
			if firstFourGameID, isFirstFour := firstFourWinners[opponentSeed]; isFirstFour {
				bracket.Games[firstFourGameID].NextGameID = gameID
				bracket.Games[firstFourGameID].NextGameSlot = 2
			} else {
				for _, t := range teamsWithByes {
					if t.Seed == opponentSeed {
						game.Team2 = b.toBracketTeam(t)
						break
					}
				}
			}
		} else {
			game.Team2 = b.toBracketTeam(team)
			if firstFourGameID, isFirstFour := firstFourWinners[opponentSeed]; isFirstFour {
				bracket.Games[firstFourGameID].NextGameID = gameID
				bracket.Games[firstFourGameID].NextGameSlot = 1
			} else {
				for _, t := range teamsWithByes {
					if t.Seed == opponentSeed {
						game.Team1 = b.toBracketTeam(t)
						break
					}
				}
			}
		}

		bracket.Games[gameID] = game
		round64Games[lowestSeed] = gameID

		processed[team.Seed] = true
		processed[opponentSeed] = true
	}

	return round64Games
}

func (b *BracketBuilder) buildRegionalRound(bracket *models.BracketStructure, region string, round models.BracketRound, previousGames map[int]string, targetSum int) map[int]string {
	newGames := make(map[int]string)

	lowestSeeds := make([]int, 0, len(previousGames))
	for lowestSeed := range previousGames {
		lowestSeeds = append(lowestSeeds, lowestSeed)
	}
	sort.Ints(lowestSeeds)

	processed := make(map[int]bool)
	gameIndex := 0
	for _, lowestSeed1 := range lowestSeeds {
		if processed[lowestSeed1] {
			continue
		}

		lowestSeed2 := targetSum - lowestSeed1
		if processed[lowestSeed2] || previousGames[lowestSeed2] == "" {
			continue
		}

		newLowestSeed := lowestSeed1
		if lowestSeed2 < lowestSeed1 {
			newLowestSeed = lowestSeed2
		}
		gameID := fmt.Sprintf("%s-%s-%d", region, round, newLowestSeed)
		game := &models.BracketGame{
			GameID:    gameID,
			Round:     round,
			Region:    region,
			SortOrder: b.getSortOrder(round, region, gameIndex),
		}

		bracket.Games[previousGames[lowestSeed1]].NextGameID = gameID
		bracket.Games[previousGames[lowestSeed1]].NextGameSlot = 1

		bracket.Games[previousGames[lowestSeed2]].NextGameID = gameID
		bracket.Games[previousGames[lowestSeed2]].NextGameSlot = 2

		bracket.Games[gameID] = game
		newGames[newLowestSeed] = gameID

		processed[lowestSeed1] = true
		processed[lowestSeed2] = true
		gameIndex++
	}

	return newGames
}

func (b *BracketBuilder) buildRegionalChampionship(bracket *models.BracketStructure, region string, elite8Games map[int]string) string {
	gameID := fmt.Sprintf("%s-elite_8-1", region)
	game := &models.BracketGame{
		GameID:    gameID,
		Round:     models.RoundElite8,
		Region:    region,
		SortOrder: b.getSortOrder(models.RoundElite8, region, 0),
	}

	seeds := make([]int, 0, len(elite8Games))
	for seed := range elite8Games {
		seeds = append(seeds, seed)
	}
	sort.Ints(seeds)

	if len(seeds) >= 2 {
		bracket.Games[elite8Games[seeds[0]]].NextGameID = gameID
		bracket.Games[elite8Games[seeds[0]]].NextGameSlot = 1

		bracket.Games[elite8Games[seeds[1]]].NextGameID = gameID
		bracket.Games[elite8Games[seeds[1]]].NextGameSlot = 2
	}

	bracket.Games[gameID] = game
	return gameID
}

func (b *BracketBuilder) buildFinalFour(bracket *models.BracketStructure, regionalChampions map[string]string, config *models.FinalFourConfig) error {
	semifinal1ID := "final_four-1"
	semifinal1 := &models.BracketGame{
		GameID:    semifinal1ID,
		Round:     models.RoundFinalFour,
		Region:    "Final Four",
		SortOrder: 1,
	}
	bracket.Games[semifinal1ID] = semifinal1

	bracket.Games[regionalChampions[config.TopLeftRegion]].NextGameID = semifinal1ID
	bracket.Games[regionalChampions[config.TopLeftRegion]].NextGameSlot = 1

	bracket.Games[regionalChampions[config.BottomLeftRegion]].NextGameID = semifinal1ID
	bracket.Games[regionalChampions[config.BottomLeftRegion]].NextGameSlot = 2

	semifinal2ID := "final_four-2"
	semifinal2 := &models.BracketGame{
		GameID:    semifinal2ID,
		Round:     models.RoundFinalFour,
		Region:    "Final Four",
		SortOrder: 2,
	}
	bracket.Games[semifinal2ID] = semifinal2

	bracket.Games[regionalChampions[config.TopRightRegion]].NextGameID = semifinal2ID
	bracket.Games[regionalChampions[config.TopRightRegion]].NextGameSlot = 1

	bracket.Games[regionalChampions[config.BottomRightRegion]].NextGameID = semifinal2ID
	bracket.Games[regionalChampions[config.BottomRightRegion]].NextGameSlot = 2

	championshipID := "championship"
	championship := &models.BracketGame{
		GameID:    championshipID,
		Round:     models.RoundChampionship,
		Region:    "Championship",
		SortOrder: 1,
	}
	bracket.Games[championshipID] = championship

	semifinal1.NextGameID = championshipID
	semifinal1.NextGameSlot = 1

	semifinal2.NextGameID = championshipID
	semifinal2.NextGameSlot = 2

	return nil
}

func (b *BracketBuilder) identifyFirstFourTeams(teams []*models.TournamentTeam) map[int][]*models.TournamentTeam {
	seedCounts := make(map[int][]*models.TournamentTeam)
	for _, team := range teams {
		seedCounts[team.Seed] = append(seedCounts[team.Seed], team)
	}

	firstFour := make(map[int][]*models.TournamentTeam)
	for seed, seedTeams := range seedCounts {
		if len(seedTeams) == 2 {
			firstFour[seed] = seedTeams
		}
	}
	return firstFour
}

func (b *BracketBuilder) getTeamsWithByes(teams []*models.TournamentTeam, firstFourTeams map[int][]*models.TournamentTeam) []*models.TournamentTeam {
	result := make([]*models.TournamentTeam, 0)
	for _, team := range teams {
		if _, isFirstFour := firstFourTeams[team.Seed]; !isFirstFour {
			result = append(result, team)
		}
	}
	return result
}

func (b *BracketBuilder) toBracketTeam(team *models.TournamentTeam) *models.BracketTeam {
	name := ""
	if team.School != nil {
		name = team.School.Name
	}
	return &models.BracketTeam{
		TeamID:         team.ID,
		SchoolID:       team.SchoolID,
		Name:           name,
		Seed:           team.Seed,
		Region:         team.Region,
		LowestSeedSeen: team.Seed,
	}
}

func (b *BracketBuilder) getSortOrder(round models.BracketRound, region string, index int) int {
	regionOrder := map[string]int{
		"East":    0,
		"West":    1,
		"South":   2,
		"Midwest": 3,
	}

	baseOrder := regionOrder[region] * 1000

	roundOrder := map[models.BracketRound]int{
		models.RoundFirstFour:    0,
		models.RoundOf64:         100,
		models.RoundOf32:         200,
		models.RoundSweet16:      300,
		models.RoundElite8:       400,
		models.RoundFinalFour:    500,
		models.RoundChampionship: 600,
	}

	return baseOrder + roundOrder[round] + index
}
