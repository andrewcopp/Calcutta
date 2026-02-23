package prediction

import (
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
)

// GenerateAllTheoreticalMatchups generates all possible matchup predictions for a 68-team tournament.
// Matchup probabilities (PMatchup) use conditional probabilities: for round R > 1,
// PMatchup(A,B) = P(A reaches this game) * P(B reaches this game), computed from
// prior-round advance probabilities.
func GenerateAllTheoreticalMatchups(teams []TeamInput, spec *simulation_game_outcomes.Spec) ([]PredictedMatchup, error) {
	if len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(teams))
	}

	if spec == nil {
		spec = &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	}
	spec.Normalize()

	// Build KenPom lookup by team ID
	kenpomByID := make(map[string]float64)
	for _, t := range teams {
		kenpomByID[t.ID] = t.KenPomNet
	}

	// Group teams by region
	teamsByRegion := make(map[string][]TeamInput)
	for _, t := range teams {
		teamsByRegion[t.Region] = append(teamsByRegion[t.Region], t)
	}

	// Get sorted region names for deterministic output
	regions := make([]string, 0, len(teamsByRegion))
	for r := range teamsByRegion {
		regions = append(regions, r)
	}
	sort.Strings(regions)

	var matchups []PredictedMatchup

	calcWinProb := func(id1, id2 string) float64 {
		net1 := kenpomByID[id1]
		net2 := kenpomByID[id2]
		return spec.WinProb(net1, net2)
	}

	// Precompute First Four win probabilities using KenPom ratings
	ffWinProb := make(map[string]float64) // teamID -> P(win FF game)
	for _, region := range regions {
		seedGroups := make(map[int][]TeamInput)
		for _, t := range teamsByRegion[region] {
			seedGroups[t.Seed] = append(seedGroups[t.Seed], t)
		}
		for _, group := range seedGroups {
			if len(group) == 2 {
				p := calcWinProb(group[0].ID, group[1].ID)
				ffWinProb[group[0].ID] = p
				ffWinProb[group[1].ID] = 1.0 - p
			}
		}
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

	// Helper to get teams by seed within a region
	teamsBySeed := func(regionTeams []TeamInput, seed int) []TeamInput {
		var result []TeamInput
		for _, t := range regionTeams {
			if t.Seed == seed {
				result = append(result, t)
			}
		}
		return result
	}

	// Helper to get teams by seeds within a region
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

	// Round 1 (Round of 64): Standard bracket matchups
	matchupSeeds := [][2]int{
		{1, 16}, {8, 9}, {5, 12}, {4, 13},
		{6, 11}, {3, 14}, {7, 10}, {2, 15},
	}

	var r1Matchups []PredictedMatchup
	for _, region := range regions {
		regionTeams := teamsByRegion[region]
		for idx, seeds := range matchupSeeds {
			t1Teams := teamsBySeed(regionTeams, seeds[0])
			t2Teams := teamsBySeed(regionTeams, seeds[1])

			for _, t1 := range t1Teams {
				for _, t2 := range t2Teams {
					// p_matchup depends on whether this is a First Four matchup
					pMatch := 1.0
					if len(t1Teams) > 1 {
						pMatch *= ffWinProb[t1.ID]
					}
					if len(t2Teams) > 1 {
						pMatch *= ffWinProb[t2.ID]
					}
					r1Matchups = append(r1Matchups, makeMatchup(
						fmt.Sprintf("R1-%s-%d", region, idx+1),
						1,
						t1, t2,
						pMatch,
					))
				}
			}
		}
	}
	matchups = append(matchups, r1Matchups...)
	pAdvance := computeAdvanceProbs(r1Matchups)

	// Round 2 (Round of 32): All possible matchups from R1 winners
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
						2,
						t1, t2,
						pMatch,
					))
				}
			}
		}
	}
	matchups = append(matchups, r2Matchups...)
	pAdvance = computeAdvanceProbs(r2Matchups)

	// Round 3 (Sweet 16): All possible matchups from R2 winners
	r3Games := []struct {
		gameNum    int
		side1Seeds []int
		side2Seeds []int
	}{
		{1, []int{1, 16, 8, 9}, []int{5, 12, 4, 13}},
		{2, []int{6, 11, 3, 14}, []int{7, 10, 2, 15}},
	}

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
						3,
						t1, t2,
						pMatch,
					))
				}
			}
		}
	}
	matchups = append(matchups, r3Matchups...)
	pAdvance = computeAdvanceProbs(r3Matchups)

	// Round 4 (Elite 8): All possible matchups from R3 winners
	topHalfSeeds := []int{1, 16, 8, 9, 5, 12, 4, 13}
	bottomHalfSeeds := []int{6, 11, 3, 14, 7, 10, 2, 15}

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
					4,
					t1, t2,
					pMatch,
				))
			}
		}
	}
	matchups = append(matchups, r4Matchups...)
	pAdvance = computeAdvanceProbs(r4Matchups)

	// Round 5 (Final Four): All possible matchups from R4 winners
	// Standard NCAA pairing: regions[0] vs regions[1], regions[2] vs regions[3]
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
					5,
					t1, t2,
					pMatch,
				))
			}
		}
	}
	matchups = append(matchups, r5Matchups...)
	pAdvance = computeAdvanceProbs(r5Matchups)

	// Round 6 (Championship): All possible matchups from R5 winners
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
				"R6-1",
				6,
				t1, t2,
				pMatch,
			))
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
