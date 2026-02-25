package bracket

import (
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func BuildBracketStructure(tournamentID string, teams []*models.TournamentTeam, finalFour *models.FinalFourConfig) (*models.BracketStructure, error) {
	if len(teams) != TotalTournamentTeams {
		return nil, fmt.Errorf("expected %d teams, got %d", TotalTournamentTeams, len(teams))
	}

	bracket := &models.BracketStructure{
		TournamentID: tournamentID,
		Regions:      append([]string{}, Regions...),
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
		championGameID, err := buildRegionalBracket(bracket, region, regionTeams)
		if err != nil {
			return nil, fmt.Errorf("error building %s region: %w", region, err)
		}
		regionalChampionGameIDs[region] = championGameID
	}

	if err := buildFinalFour(bracket, regionalChampionGameIDs, finalFour); err != nil {
		return nil, fmt.Errorf("error building Final Four: %w", err)
	}

	return bracket, nil
}

func buildRegionalBracket(bracket *models.BracketStructure, region string, teams []*models.TournamentTeam) (string, error) {
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Seed < teams[j].Seed
	})

	firstFourTeams := identifyFirstFourTeams(teams)

	firstFourWinners := make(map[int]string)
	for seed, seedTeams := range firstFourTeams {
		if len(seedTeams) == 2 {
			gameID := fmt.Sprintf("%s-first_four-%d", region, seed)
			game := &models.BracketGame{
				GameID:    gameID,
				Round:     models.RoundFirstFour,
				Region:    region,
				Team1:     toBracketTeam(seedTeams[0]),
				Team2:     toBracketTeam(seedTeams[1]),
				SortOrder: getSortOrder(models.RoundFirstFour, region, seed),
			}
			bracket.Games[gameID] = game
			firstFourWinners[seed] = gameID
		}
	}

	round64Games := buildRoundOf64(bracket, region, teams, firstFourWinners)
	round32Games := buildRegionalRound(bracket, region, models.RoundOf32, round64Games, SeedPairSumR32)
	sweet16Games := buildRegionalRound(bracket, region, models.RoundSweet16, round32Games, SeedPairSumS16)
	elite8Games := buildRegionalRound(bracket, region, models.RoundElite8, sweet16Games, SeedPairSumE8)
	championGameID := buildRegionalChampionship(bracket, region, elite8Games)

	return championGameID, nil
}

func buildRoundOf64(bracket *models.BracketStructure, region string, teams []*models.TournamentTeam, firstFourWinners map[int]string) map[int]string {
	round64Games := make(map[int]string)
	teamsWithByes := getTeamsWithByes(teams, identifyFirstFourTeams(teams))

	processed := make(map[int]bool)
	for _, team := range teamsWithByes {
		if processed[team.Seed] {
			continue
		}

		opponentSeed := SeedPairSumR64 - team.Seed
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
			SortOrder: getSortOrder(models.RoundOf64, region, team.Seed),
		}

		if team.Seed < opponentSeed {
			game.Team1 = toBracketTeam(team)
			if firstFourGameID, isFirstFour := firstFourWinners[opponentSeed]; isFirstFour {
				bracket.Games[firstFourGameID].NextGameID = gameID
				bracket.Games[firstFourGameID].NextGameSlot = 2
			} else {
				for _, t := range teamsWithByes {
					if t.Seed == opponentSeed {
						game.Team2 = toBracketTeam(t)
						break
					}
				}
			}
		} else {
			game.Team2 = toBracketTeam(team)
			if firstFourGameID, isFirstFour := firstFourWinners[opponentSeed]; isFirstFour {
				bracket.Games[firstFourGameID].NextGameID = gameID
				bracket.Games[firstFourGameID].NextGameSlot = 1
			} else {
				for _, t := range teamsWithByes {
					if t.Seed == opponentSeed {
						game.Team1 = toBracketTeam(t)
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

func buildRegionalRound(bracket *models.BracketStructure, region string, round models.BracketRound, previousGames map[int]string, targetSum int) map[int]string {
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
			SortOrder: getSortOrder(round, region, gameIndex),
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

func buildRegionalChampionship(bracket *models.BracketStructure, region string, elite8Games map[int]string) string {
	gameID := fmt.Sprintf("%s-elite_8-1", region)
	game := &models.BracketGame{
		GameID:    gameID,
		Round:     models.RoundElite8,
		Region:    region,
		SortOrder: getSortOrder(models.RoundElite8, region, 0),
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

func buildFinalFour(bracket *models.BracketStructure, regionalChampions map[string]string, config *models.FinalFourConfig) error {
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

func identifyFirstFourTeams(teams []*models.TournamentTeam) map[int][]*models.TournamentTeam {
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

func getTeamsWithByes(teams []*models.TournamentTeam, firstFourTeams map[int][]*models.TournamentTeam) []*models.TournamentTeam {
	result := make([]*models.TournamentTeam, 0)
	for _, team := range teams {
		if _, isFirstFour := firstFourTeams[team.Seed]; !isFirstFour {
			result = append(result, team)
		}
	}
	return result
}

func toBracketTeam(team *models.TournamentTeam) *models.BracketTeam {
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

func getSortOrder(round models.BracketRound, region string, index int) int {
	return regionSortOrder[region]*RegionSortMultiplier + roundSortOffset[round] + index
}
