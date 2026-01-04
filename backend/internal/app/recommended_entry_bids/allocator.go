package recommended_entry_bids

import "sort"

type Team struct {
	ID             string
	ExpectedPoints float64
	MarketPoints   float64
}

type AllocationParams struct {
	BudgetPoints int
	MinTeams     int
	MaxTeams     int
	MinBidPoints int
	MaxBidPoints int
}

type AllocationResult struct {
	Bids map[string]int
}

func AllocateBids(teams []Team, params AllocationParams) (AllocationResult, error) {
	if params.BudgetPoints <= 0 {
		return AllocationResult{Bids: map[string]int{}}, nil
	}
	if params.MinTeams <= 0 {
		params.MinTeams = 1
	}
	if params.MaxTeams < params.MinTeams {
		params.MaxTeams = params.MinTeams
	}
	if params.MinBidPoints <= 0 {
		params.MinBidPoints = 1
	}
	if params.MaxBidPoints < params.MinBidPoints {
		params.MaxBidPoints = params.MinBidPoints
	}

	required := params.MinTeams * params.MinBidPoints
	if required > params.BudgetPoints {
		return AllocationResult{Bids: map[string]int{}}, nil
	}

	sorted := make([]Team, 0, len(teams))
	for _, t := range teams {
		sorted = append(sorted, t)
	}
	sort.Slice(sorted, func(i, j int) bool {
		ai := objective(sorted[i], params.MinBidPoints)
		aj := objective(sorted[j], params.MinBidPoints)
		if ai != aj {
			return ai > aj
		}
		return sorted[i].ID < sorted[j].ID
	})

	selected := make(map[string]int)
	remaining := params.BudgetPoints
	for i := 0; i < len(sorted) && len(selected) < params.MinTeams; i++ {
		selected[sorted[i].ID] = params.MinBidPoints
		remaining -= params.MinBidPoints
	}

	if remaining < 0 {
		return AllocationResult{Bids: map[string]int{}}, nil
	}

	byID := make(map[string]Team, len(sorted))
	for _, t := range sorted {
		byID[t.ID] = t
	}

	for remaining > 0 {
		selectedIDs := make([]string, 0, len(selected))
		for teamID := range selected {
			selectedIDs = append(selectedIDs, teamID)
		}
		sort.Strings(selectedIDs)

		bestGainPerPoint := -1.0
		bestTeamID := ""
		bestIsNew := false
		bestCost := 0

		for _, teamID := range selectedIDs {
			bid := selected[teamID]
			if bid >= params.MaxBidPoints {
				continue
			}
			gain := marginalGain(byID[teamID], bid)
			if gain > bestGainPerPoint || (gain == bestGainPerPoint && teamID < bestTeamID) {
				bestGainPerPoint = gain
				bestTeamID = teamID
				bestIsNew = false
				bestCost = 1
			}
		}

		if len(selected) < params.MaxTeams && remaining >= params.MinBidPoints {
			for _, t := range sorted {
				if _, ok := selected[t.ID]; ok {
					continue
				}
				gain := objective(t, params.MinBidPoints)
				gainPerPoint := gain / float64(params.MinBidPoints)
				if gainPerPoint > bestGainPerPoint || (gainPerPoint == bestGainPerPoint && t.ID < bestTeamID) {
					bestGainPerPoint = gainPerPoint
					bestTeamID = t.ID
					bestIsNew = true
					bestCost = params.MinBidPoints
				}
			}
		}

		if bestTeamID == "" {
			break
		}
		if bestIsNew {
			selected[bestTeamID] = params.MinBidPoints
			remaining -= bestCost
			continue
		}
		selected[bestTeamID]++
		remaining -= bestCost
	}

	return AllocationResult{Bids: selected}, nil
}

func objective(t Team, bid int) float64 {
	if bid <= 0 {
		return 0
	}
	den := t.MarketPoints + float64(bid)
	if den <= 0 {
		return 0
	}
	return t.ExpectedPoints * float64(bid) / den
}

func marginalGain(t Team, bid int) float64 {
	return objective(t, bid+1) - objective(t, bid)
}
