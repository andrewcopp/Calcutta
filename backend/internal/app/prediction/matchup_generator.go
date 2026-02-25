package prediction

import (
	"fmt"
	"sort"
	"strings"

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

// byePrefix is the sentinel prefix for phantom BYE opponents in the R128 model.
const byePrefix = "BYE-"

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
// Rounds 2-5 correspond to R64, R32, S16, E8 in the 128-team symmetric model.
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
	case 2: // R64
		games := make([]gameSetup, 0, len(r64Pairings))
		for idx, seeds := range r64Pairings {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R2-%s-%d", region, idx+1),
				side1:  teamsBySeed(seeds[0]),
				side2:  teamsBySeed(seeds[1]),
			})
		}
		return games
	case 3: // R32
		games := make([]gameSetup, 0, len(r32Groups))
		for idx, group := range r32Groups {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R3-%s-%d", region, idx+1),
				side1:  teamsBySeeds(group[0]),
				side2:  teamsBySeeds(group[1]),
			})
		}
		return games
	case 4: // S16
		games := make([]gameSetup, 0, len(s16Groups))
		for idx, group := range s16Groups {
			games = append(games, gameSetup{
				gameID: fmt.Sprintf("R4-%s-%d", region, idx+1),
				side1:  teamsBySeeds(group[0]),
				side2:  teamsBySeeds(group[1]),
			})
		}
		return games
	case 5: // E8
		return []gameSetup{{
			gameID: fmt.Sprintf("R5-%s-1", region),
			side1:  teamsBySeeds(topHalfSeeds),
			side2:  teamsBySeeds(bottomHalfSeeds),
		}}
	}
	return nil
}

// buildR128Games creates Round 1 game setups for the 128-team symmetric model.
// Seeds with 1 team get a phantom BYE opponent (guaranteed win).
// Seeds with 2 teams (First Four) play a real game.
// Returns the game setups and any BYE sentinel TeamInputs that were created.
func buildR128Games(region string, regionTeams []TeamInput) ([]gameSetup, []TeamInput) {
	seedGroups := make(map[int][]TeamInput)
	for _, t := range regionTeams {
		seedGroups[t.Seed] = append(seedGroups[t.Seed], t)
	}

	var games []gameSetup
	var byeTeams []TeamInput
	gameIdx := 1

	// Process seeds in bracket order (matching r64Pairings) so each R128 game
	// feeds into the correct R64 slot.
	for _, pairing := range r64Pairings {
		for _, seed := range []int{pairing[0], pairing[1]} {
			group := seedGroups[seed]
			if len(group) == 2 {
				// Real First Four game.
				games = append(games, gameSetup{
					gameID: fmt.Sprintf("R1-%s-%d", region, gameIdx),
					side1:  []TeamInput{group[0]},
					side2:  []TeamInput{group[1]},
				})
			} else if len(group) == 1 {
				// Phantom BYE game: real team vs BYE sentinel.
				byeID := fmt.Sprintf("%s%s-%d", byePrefix, region, seed)
				bye := TeamInput{ID: byeID, Seed: seed, Region: region}
				byeTeams = append(byeTeams, bye)
				games = append(games, gameSetup{
					gameID: fmt.Sprintf("R1-%s-%d", region, gameIdx),
					side1:  []TeamInput{group[0]},
					side2:  []TeamInput{bye},
				})
			}
			gameIdx++
		}
	}

	return games, byeTeams
}

// GenerateMatchups generates matchup predictions starting from a tournament checkpoint.
// The tournament is modeled as a 128-team symmetric binary tree where 60 phantom BYE
// opponents guarantee Round 1 wins, making all 7 rounds structurally identical.
//
// For throughRound == 0 (pre-tournament), it requires exactly 68 teams and creates
// BYE sentinels to fill the R128 bracket.
// For throughRound >= 1, all survivors start with pAdvance = 1.0.
// Rounds already resolved (<= throughRound) are skipped.
// ffConfig determines the Final Four region pairings; if nil, defaults are applied.
func GenerateMatchups(teams []TeamInput, throughRound int, spec *simulation_game_outcomes.Spec, ffConfig *models.FinalFourConfig) ([]PredictedMatchup, error) {
	if throughRound == 0 && len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams for pre-tournament predictions, got %d", len(teams))
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
		// BYE opponents always lose.
		if strings.HasPrefix(id2, byePrefix) {
			return 1.0
		}
		if strings.HasPrefix(id1, byePrefix) {
			return 0.0
		}
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

	// Initialize pAdvance for all real teams.
	pAdvance := make(map[string]float64, len(teams))
	for _, t := range teams {
		pAdvance[t.ID] = 1.0
	}

	var matchups []PredictedMatchup

	// Round 1 (R128): phantom BYE games + real First Four games.
	if throughRound < 1 {
		var r128Games []gameSetup
		for _, region := range regions {
			games, byeTeams := buildR128Games(region, teamsByRegion[region])
			r128Games = append(r128Games, games...)
			// Add BYE sentinels to pAdvance with probability 1.0.
			for _, bye := range byeTeams {
				pAdvance[bye.ID] = 1.0
			}
		}
		var roundMatchups []PredictedMatchup
		roundMatchups, pAdvance = computeRound(r128Games, pAdvance, calcWinProb, 1)
		matchups = append(matchups, roundMatchups...)
	}

	// Rounds 2-5: regional rounds (R64, R32, S16, E8).
	for round := 2; round <= 5; round++ {
		if throughRound >= round {
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

	// Round 6 (Final Four semifinals): use ffConfig for pairings.
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
					gameID: fmt.Sprintf("R6-%d", gameNum+1),
					side1:  teams1,
					side2:  teams2,
				})
			}
		}
		if len(games) > 0 {
			var roundMatchups []PredictedMatchup
			roundMatchups, pAdvance = computeRound(games, pAdvance, calcWinProb, 6)
			matchups = append(matchups, roundMatchups...)
		}
	}

	// Round 7 (Championship): left bracket vs right bracket.
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
				gameID: "R7-1",
				side1:  side1Teams,
				side2:  side2Teams,
			}}
			roundMatchups, _ := computeRound(games, pAdvance, calcWinProb, 7)
			matchups = append(matchups, roundMatchups...)
		}
	}

	return matchups, nil
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
