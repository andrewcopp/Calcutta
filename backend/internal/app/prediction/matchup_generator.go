package prediction

import (
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

// GenerateMatchups generates matchup predictions starting from a tournament checkpoint.
// For throughRound == 0 (pre-tournament), it computes First Four win probabilities using
// KenPom ratings and requires exactly 68 teams.
// For throughRound >= 1, all survivors start with pAdvance = 1.0.
// Rounds already resolved (< throughRound+1) are skipped.
func GenerateMatchups(teams []TeamInput, throughRound int, spec *simulation_game_outcomes.Spec) ([]PredictedMatchup, error) {
	if throughRound == 0 && len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams for pre-tournament predictions, got %d", len(teams))
	}

	if spec == nil {
		spec = &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	}
	spec.Normalize()

	kenpomByID := make(map[string]float64)
	for _, t := range teams {
		kenpomByID[t.ID] = t.KenPomNet
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

	calcWinProb := func(id1, id2 string) float64 {
		return spec.WinProb(kenpomByID[id1], kenpomByID[id2])
	}

	makeMatchup := func(gameID string, roundOrder int, t1, t2 TeamInput, pMatchup float64) PredictedMatchup {
		p1Wins := calcWinProb(t1.ID, t2.ID)
		return PredictedMatchup{
			GameID:                 gameID,
			RoundOrder:             roundOrder,
			Team1ID:                t1.ID,
			Team2ID:                t2.ID,
			PMatchup:               pMatchup,
			PTeam1WinsGivenMatchup: p1Wins,
			PTeam2WinsGivenMatchup: 1.0 - p1Wins,
		}
	}

	teamsBySeed := func(regionTeams []TeamInput, seed int) []TeamInput {
		var result []TeamInput
		for _, t := range regionTeams {
			if t.Seed == seed {
				result = append(result, t)
			}
		}
		return result
	}

	teamsBySeeds := func(regionTeams []TeamInput, seeds []int) []TeamInput {
		seedSet := make(map[int]bool)
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

	matchupSeeds := [][2]int{
		{1, 16}, {8, 9}, {5, 12}, {4, 13},
		{6, 11}, {3, 14}, {7, 10}, {2, 15},
	}

	// Initialize pAdvance for all teams
	pAdvance := make(map[string]float64)
	if throughRound < 1 {
		// Pre-tournament: compute First Four win probabilities
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
		// Mid-tournament: all survivors have pAdvance = 1.0
		for _, t := range teams {
			pAdvance[t.ID] = 1.0
		}
	}

	var matchups []PredictedMatchup

	// Round 1 (R64): only if throughRound < 2
	if throughRound < 2 {
		var r1Matchups []PredictedMatchup
		for _, region := range regions {
			regionTeams := teamsByRegion[region]
			for idx, seeds := range matchupSeeds {
				t1Teams := teamsBySeed(regionTeams, seeds[0])
				t2Teams := teamsBySeed(regionTeams, seeds[1])
				for _, t1 := range t1Teams {
					for _, t2 := range t2Teams {
						pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
						r1Matchups = append(r1Matchups, makeMatchup(
							fmt.Sprintf("R1-%s-%d", region, idx+1),
							1, t1, t2, pMatch,
						))
					}
				}
			}
		}
		matchups = append(matchups, r1Matchups...)
		pAdvance = computeAdvanceProbs(r1Matchups)
	}

	r2Games := []struct {
		gameNum    int
		side1Seeds []int
		side2Seeds []int
	}{
		{1, []int{1, 16}, []int{8, 9}},
		{2, []int{5, 12}, []int{4, 13}},
		{3, []int{6, 11}, []int{3, 14}},
		{4, []int{7, 10}, []int{2, 15}},
	}

	// Round 2 (R32): only if throughRound < 3
	if throughRound < 3 {
		var r2Matchups []PredictedMatchup
		for _, region := range regions {
			regionTeams := teamsByRegion[region]
			for _, game := range r2Games {
				side1 := teamsBySeeds(regionTeams, game.side1Seeds)
				side2 := teamsBySeeds(regionTeams, game.side2Seeds)
				for _, t1 := range side1 {
					for _, t2 := range side2 {
						pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
						r2Matchups = append(r2Matchups, makeMatchup(
							fmt.Sprintf("R2-%s-%d", region, game.gameNum),
							2, t1, t2, pMatch,
						))
					}
				}
			}
		}
		matchups = append(matchups, r2Matchups...)
		pAdvance = computeAdvanceProbs(r2Matchups)
	}

	r3Games := []struct {
		gameNum    int
		side1Seeds []int
		side2Seeds []int
	}{
		{1, []int{1, 16, 8, 9}, []int{5, 12, 4, 13}},
		{2, []int{6, 11, 3, 14}, []int{7, 10, 2, 15}},
	}

	// Round 3 (S16): only if throughRound < 4
	if throughRound < 4 {
		var r3Matchups []PredictedMatchup
		for _, region := range regions {
			regionTeams := teamsByRegion[region]
			for _, game := range r3Games {
				side1 := teamsBySeeds(regionTeams, game.side1Seeds)
				side2 := teamsBySeeds(regionTeams, game.side2Seeds)
				for _, t1 := range side1 {
					for _, t2 := range side2 {
						pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
						r3Matchups = append(r3Matchups, makeMatchup(
							fmt.Sprintf("R3-%s-%d", region, game.gameNum),
							3, t1, t2, pMatch,
						))
					}
				}
			}
		}
		matchups = append(matchups, r3Matchups...)
		pAdvance = computeAdvanceProbs(r3Matchups)
	}

	topHalfSeeds := []int{1, 16, 8, 9, 5, 12, 4, 13}
	bottomHalfSeeds := []int{6, 11, 3, 14, 7, 10, 2, 15}

	// Round 4 (E8): only if throughRound < 5
	if throughRound < 5 {
		var r4Matchups []PredictedMatchup
		for _, region := range regions {
			regionTeams := teamsByRegion[region]
			side1 := teamsBySeeds(regionTeams, topHalfSeeds)
			side2 := teamsBySeeds(regionTeams, bottomHalfSeeds)
			for _, t1 := range side1 {
				for _, t2 := range side2 {
					pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
					r4Matchups = append(r4Matchups, makeMatchup(
						fmt.Sprintf("R4-%s-1", region),
						4, t1, t2, pMatch,
					))
				}
			}
		}
		matchups = append(matchups, r4Matchups...)
		pAdvance = computeAdvanceProbs(r4Matchups)
	}

	// Round 5 (F4): only if throughRound < 6 and we have 4 regions for bracket pairing
	if throughRound < 6 && len(regions) >= 4 {
		ffPairings := [][2]string{
			{regions[0], regions[1]},
			{regions[2], regions[3]},
		}
		var r5Matchups []PredictedMatchup
		for gameNum, pairing := range ffPairings {
			teams1 := teamsByRegion[pairing[0]]
			teams2 := teamsByRegion[pairing[1]]
			for _, t1 := range teams1 {
				for _, t2 := range teams2 {
					pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
					r5Matchups = append(r5Matchups, makeMatchup(
						fmt.Sprintf("R5-%d", gameNum+1),
						5, t1, t2, pMatch,
					))
				}
			}
		}
		matchups = append(matchups, r5Matchups...)
		pAdvance = computeAdvanceProbs(r5Matchups)
	}

	// Round 6 (NCG): only if throughRound < 7 and we have 4 regions for bracket pairing
	if throughRound < 7 && len(regions) >= 4 {
		side1Regions := []string{regions[0], regions[1]}
		side2Regions := []string{regions[2], regions[3]}

		var side1Teams, side2Teams []TeamInput
		for _, r := range side1Regions {
			side1Teams = append(side1Teams, teamsByRegion[r]...)
		}
		for _, r := range side2Regions {
			side2Teams = append(side2Teams, teamsByRegion[r]...)
		}

		for _, t1 := range side1Teams {
			for _, t2 := range side2Teams {
				pMatch := pAdvance[t1.ID] * pAdvance[t2.ID]
				matchups = append(matchups, makeMatchup(
					"R6-1", 6, t1, t2, pMatch,
				))
			}
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
