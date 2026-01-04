package recommended_entry_bids

import (
	"math"
	"sort"
)

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
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].ID < teams[j].ID
	})

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

	nTeams := len(teams)
	maxBudget := params.BudgetPoints
	maxTeams := params.MaxTeams

	negInf := math.Inf(-1)

	dpPrev := make([][]float64, maxBudget+1)
	dpCur := make([][]float64, maxBudget+1)
	for b := 0; b <= maxBudget; b++ {
		dpPrev[b] = make([]float64, maxTeams+1)
		dpCur[b] = make([]float64, maxTeams+1)
		for k := 0; k <= maxTeams; k++ {
			dpPrev[b][k] = negInf
			dpCur[b][k] = negInf
		}
	}
	dpPrev[0][0] = 0.0

	prevBudget := make([][][]int, nTeams+1)
	prevCount := make([][][]int, nTeams+1)
	chosenBid := make([][][]int, nTeams+1)
	for i := 0; i <= nTeams; i++ {
		prevBudget[i] = make([][]int, maxBudget+1)
		prevCount[i] = make([][]int, maxBudget+1)
		chosenBid[i] = make([][]int, maxBudget+1)
		for b := 0; b <= maxBudget; b++ {
			prevBudget[i][b] = make([]int, maxTeams+1)
			prevCount[i][b] = make([]int, maxTeams+1)
			chosenBid[i][b] = make([]int, maxTeams+1)
			for k := 0; k <= maxTeams; k++ {
				prevBudget[i][b][k] = -1
				prevCount[i][b][k] = -1
				chosenBid[i][b][k] = -1
			}
		}
	}

	for i := 0; i < nTeams; i++ {
		for b := 0; b <= maxBudget; b++ {
			for k := 0; k <= maxTeams; k++ {
				dpCur[b][k] = negInf
			}
		}

		t := teams[i]

		options := make([]int, 0, (params.MaxBidPoints-params.MinBidPoints+1)+1)
		options = append(options, 0)
		for bid := params.MinBidPoints; bid <= params.MaxBidPoints; bid++ {
			options = append(options, bid)
		}

		for b := 0; b <= maxBudget; b++ {
			for k := 0; k <= maxTeams; k++ {
				base := dpPrev[b][k]
				if math.IsInf(base, -1) {
					continue
				}
				for _, bid := range options {
					cost := bid
					if b+cost > maxBudget {
						continue
					}
					k2 := k
					if bid > 0 {
						k2 = k + 1
					}
					if k2 > maxTeams {
						continue
					}

					v := base + objective(t, bid)
					cur := dpCur[b+cost][k2]
					if v > cur {
						dpCur[b+cost][k2] = v
						prevBudget[i+1][b+cost][k2] = b
						prevCount[i+1][b+cost][k2] = k
						chosenBid[i+1][b+cost][k2] = bid
						continue
					}
					if v == cur {
						prevBid := chosenBid[i+1][b+cost][k2]
						if prevBid < 0 || bid < prevBid {
							prevBudget[i+1][b+cost][k2] = b
							prevCount[i+1][b+cost][k2] = k
							chosenBid[i+1][b+cost][k2] = bid
						}
					}
				}
			}
		}

		dpPrev, dpCur = dpCur, dpPrev
	}

	bestV := negInf
	bestBudget := -1
	bestTeams := -1
	for b := 0; b <= maxBudget; b++ {
		for k := params.MinTeams; k <= maxTeams; k++ {
			v := dpPrev[b][k]
			if math.IsInf(v, -1) {
				continue
			}
			if v > bestV {
				bestV = v
				bestBudget = b
				bestTeams = k
				continue
			}
			if v == bestV {
				if b > bestBudget {
					bestBudget = b
					bestTeams = k
					continue
				}
				if b == bestBudget && k < bestTeams {
					bestTeams = k
				}
			}
		}
	}
	if bestBudget < 0 || bestTeams < 0 || math.IsInf(bestV, -1) {
		return AllocationResult{Bids: map[string]int{}}, nil
	}

	bids := make(map[string]int)
	b := bestBudget
	k := bestTeams
	for i := nTeams; i >= 1; i-- {
		bid := chosenBid[i][b][k]
		if bid < 0 {
			break
		}
		if bid > 0 {
			teamID := teams[i-1].ID
			bids[teamID] = bid
		}
		pb := prevBudget[i][b][k]
		pk := prevCount[i][b][k]
		if pb < 0 || pk < 0 {
			break
		}
		b = pb
		k = pk
	}

	return AllocationResult{Bids: bids}, nil
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
