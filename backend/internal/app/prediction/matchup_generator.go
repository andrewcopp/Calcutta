package prediction

import (
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// gameSetup describes one bracket game slot and the candidate teams on each side.
type gameSetup struct {
	gameID string
	side1  []TeamInput
	side2  []TeamInput
}

// Standard NCAA bracket seed pairings for the Round of 64.
var r64Pairings = [][2]int{
	{1, 16}, {8, 9}, {5, 12}, {4, 13},
	{6, 11}, {3, 14}, {7, 10}, {2, 15},
}

// r32Groups defines which seed groups face off in the Round of 32.
var r32Groups = [][2][]int{
	{{1, 16}, {8, 9}},
	{{5, 12}, {4, 13}},
	{{6, 11}, {3, 14}},
	{{7, 10}, {2, 15}},
}

// s16Groups defines which seed groups face off in the Sweet 16.
var s16Groups = [][2][]int{
	{{1, 16, 8, 9}, {5, 12, 4, 13}},
	{{6, 11, 3, 14}, {7, 10, 2, 15}},
}

// topHalfSeeds and bottomHalfSeeds define the Elite 8 matchup within a region.
var topHalfSeeds = []int{1, 16, 8, 9, 5, 12, 4, 13}
var bottomHalfSeeds = []int{6, 11, 3, 14, 7, 10, 2, 15}

// computeRound generates matchups for one round from the given game setups
// and returns both the matchups and updated advance probabilities.
func computeRound(
	games []gameSetup,
	pAdvance map[string]float64,
	calcWinProb func(id1, id2 string) float64,
	roundOrder int,
) ([]PredictedMatchup, map[string]float64) {
	var matchups []PredictedMatchup
	for _, g := range games {
		for _, t1 := range g.side1 {
			for _, t2 := range g.side2 {
				pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
				p1Wins := calcWinProb(t1.ID, t2.ID)
				matchups = append(matchups, PredictedMatchup{
					GameID:                 g.gameID,
					RoundOrder:             roundOrder,
					Team1ID:                t1.ID,
					Team2ID:                t2.ID,
					PMatchup:               pMatch,
					PTeam1WinsGivenMatchup: p1Wins,
					PTeam2WinsGivenMatchup: 1.0 - p1Wins,
				})
			}
		}
	}
	return matchups, computeAdvanceProbs(matchups)
}

// buildRegionalGames creates game setups for a single region at a given round.
func buildRegionalGames(region string, regionTeams []TeamInput, round int) []gameSetup {
	teamsBySeed := func(seed int) []TeamInput {
		var result []TeamInput
		for _, t := range regionTeams {
			if t.Seed == seed {
				result = append(result, t)
			}
		}
		return result
	}

	teamsBySeeds := func(seeds []int) []TeamInput {
		seedSet := make(map[int]bool, len(seeds))
		for _, s := range seeds {
			seedSet[s] = true
		}
		var result []TeamInput
		for _, t := range regionTeams {
			if seedSet[t.Seed] {
				result = append(result, t)
			}
		}
		return result
	}

	switch round {
	case 1: // R64
		games := make([]gameSetup, 0, len(r64Pairings))
		for idx, seeds := range r64Pairings {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R1-%s-%d", region, idx+1),
				side1:  teamsBySeed(seeds[0]),
				side2:  teamsBySeed(seeds[1]),
			})
		}
		return games
	case 2: // R32
		games := make([]gameSetup, 0, len(r32Groups))
		for idx, group := range r32Groups {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R2-%s-%d", region, idx+1),
				side1:  teamsBySeeds(group[0]),
				side2:  teamsBySeeds(group[1]),
			})
		}
		return games
	case 3: // S16
		games := make([]gameSetup, 0, len(s16Groups))
		for idx, group := range s16Groups {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R3-%s-%d", region, idx+1),
				side1:  teamsBySeeds(group[0]),
				side2:  teamsBySeeds(group[1]),
			})
		}
		return games
	case 4: // E8
		return []gameSetup{{
			gameID: fmt.Sprintf("R4-%s-1", region),
			side1:  teamsBySeeds(topHalfSeeds),
			side2:  teamsBySeeds(bottomHalfSeeds),
		}}
	}
	return nil
}

// GenerateMatchups generates matchup predictions starting from a tournament checkpoint.
// For throughRound == 0 (pre-tournament), it computes First Four win probabilities using
// KenPom ratings and requires exactly 68 teams.
// For throughRound >= 1, all survivors start with pAdvance = 1.0.
// Rounds already resolved (< throughRound+1) are skipped.
// ffConfig determines the Final Four region pairings; if nil, defaults are applied.
func GenerateMatchups(teams []TeamInput, throughRound int, spec *simulation_game_outcomes.Spec, ffConfig *models.FinalFourConfig) ([]PredictedMatchup, map[string]float64, error) {
	if throughRound == 0 && len(teams) != 68 {
		return nil, nil, fmt.Errorf("expected 68 teams for pre-tournament predictions, got %d", len(teams))
	}

	if spec == nil {
		spec = &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	}
	spec.Normalize()

	if ffConfig == nil {
		ffConfig = &models.FinalFourConfig{}
		ffConfig.ApplyDefaults()
	}

	kenpomByID := make(map[string]float64, len(teams))
	for _, t := range teams {
		kenpomByID[t.ID] = t.KenPomNet
	}

	calcWinProb := func(id1, id2 string) float64 {
		return spec.WinProb(kenpomByID[id1], kenpomByID[id2])
	}

	teamsByRegion := make(map[string][]TeamInput)
	for _, t := range teams {
		teamsByRegion[t.Region] = append(teamsByRegion[t.Region], t)
	}

	regions := make([]string, 0, len(teamsByRegion))
	for r := range teamsByRegion {
		regions = append(regions, r)
	}
	sort.Strings(regions)

	// Initialize pAdvance for all teams.
	pAdvance := make(map[string]float64, len(teams))
	if throughRound < 1 {
		// Pre-tournament: compute First Four win probabilities.
		for _, t := range teams {
			pAdvance[t.ID] = 1.0
		}
		for _, region := range regions {
			seedGroups := make(map[int][]TeamInput)
			for _, t := range teamsByRegion[region] {
				seedGroups[t.Seed] = append(seedGroups[t.Seed], t)
			}
			for _, group := range seedGroups {
				if len(group) == 2 {
					p := calcWinProb(group[0].ID, group[1].ID)
					pAdvance[group[0].ID] = p
					pAdvance[group[1].ID] = 1.0 - p
				}
			}
		}
	} else {
		// Mid-tournament: all survivors have pAdvance = 1.0.
		for _, t := range teams {
			pAdvance[t.ID] = 1.0
		}
	}

	// Capture play-in survival probabilities before the round loop modifies pAdvance.
	// For pre-tournament: 1.0 for bye teams, <1.0 for FF teams.
	// For mid-tournament: all 1.0 (play-in already resolved).
	pPlayinSurvival := make(map[string]float64, len(pAdvance))
	for id, p := range pAdvance {
		pPlayinSurvival[id] = p
	}

	var matchups []PredictedMatchup

	// Rounds 1-4: regional rounds (R64, R32, S16, E8).
	for round := 1; round <= 4; round++ {
		if throughRound >= round+1 {
			continue
		}
		var games []gameSetup
		for _, region := range regions {
			games = append(games, buildRegionalGames(region, teamsByRegion[region], round)...)
		}
		var roundMatchups []PredictedMatchup
		roundMatchups, pAdvance = computeRound(games, pAdvance, calcWinProb, round)
		matchups = append(matchups, roundMatchups...)
	}

	// Round 5 (Final Four semifinals): use ffConfig for pairings.
	if throughRound < 6 {
		ffPairings := [][2]string{
			{ffConfig.TopLeftRegion, ffConfig.BottomLeftRegion},
			{ffConfig.TopRightRegion, ffConfig.BottomRightRegion},
		}
		var games []gameSetup
		for gameNum, pairing := range ffPairings {
			teams1 := teamsByRegion[pairing[0]]
			teams2 := teamsByRegion[pairing[1]]
			if len(teams1) > 0 && len(teams2) > 0 {
				games = append(games, gameSetup{
					gameID: fmt.Sprintf("R5-%d", gameNum+1),
					side1:  teams1,
					side2:  teams2,
				})
			}
		}
		if len(games) > 0 {
			var roundMatchups []PredictedMatchup
			roundMatchups, pAdvance = computeRound(games, pAdvance, calcWinProb, 5)
			matchups = append(matchups, roundMatchups...)
		}
	}

	// Round 6 (Championship): left bracket vs right bracket.
	if throughRound < 7 {
		var side1Teams, side2Teams []TeamInput
		for _, r := range []string{ffConfig.TopLeftRegion, ffConfig.BottomLeftRegion} {
			side1Teams = append(side1Teams, teamsByRegion[r]...)
		}
		for _, r := range []string{ffConfig.TopRightRegion, ffConfig.BottomRightRegion} {
			side2Teams = append(side2Teams, teamsByRegion[r]...)
		}
		if len(side1Teams) > 0 && len(side2Teams) > 0 {
			games := []gameSetup{{
				gameID: "R6-1",
				side1:  side1Teams,
				side2:  side2Teams,
			}}
			roundMatchups, _ := computeRound(games, pAdvance, calcWinProb, 6)
			matchups = append(matchups, roundMatchups...)
		}
	}

	return matchups, pPlayinSurvival, nil
}

// computeAdvanceProbs computes each team's probability of advancing past the given round.
// For each matchup, a team's advance probability is pMatchup * pWin, summed across all
// matchups involving that team in the round.
func computeAdvanceProbs(roundMatchups []PredictedMatchup) map[string]float64 {
	pAdvance := make(map[string]float64)
	for _, m := range roundMatchups {
		pAdvance[m.Team1ID] += m.PMatchup * m.PTeam1WinsGivenMatchup
		pAdvance[m.Team2ID] += m.PMatchup * m.PTeam2WinsGivenMatchup
	}
	return pAdvance
}
