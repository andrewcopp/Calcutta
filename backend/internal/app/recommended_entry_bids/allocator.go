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

// dpTables holds the dynamic programming state and backtracking pointers
// used during bid allocation optimization.
type dpTables struct {
	dpPrev     [][]float64
	dpCur      [][]float64
	prevBudget [][][]int
	prevCount  [][][]int
	chosenBid  [][][]int
	maxBudget  int
	maxTeams   int
}

func AllocateBids(teams []Team, params AllocationParams) (AllocationResult, error) {
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].ID < teams[j].ID
	})

	params = normalizeParams(params)

	if params.BudgetPoints <= 0 {
		return AllocationResult{Bids: map[string]int{}}, nil
	}
	if params.MinTeams*params.MinBidPoints > params.BudgetPoints {
		return AllocationResult{Bids: map[string]int{}}, nil
	}

	nTeams := len(teams)
	dp := initDPTables(nTeams, params.BudgetPoints, params.MaxTeams)

	solveDPTransitions(dp, teams, params)

	bestBudget, bestTeams, found := selectBestSolution(dp, params)
	if !found {
		return AllocationResult{Bids: map[string]int{}}, nil
	}

	bids := backtrackBids(dp, teams, bestBudget, bestTeams)
	return AllocationResult{Bids: bids}, nil
}

// normalizeParams ensures all allocation parameters have valid minimums
// and consistent relationships (e.g., max >= min).
func normalizeParams(p AllocationParams) AllocationParams {
	if p.MinTeams <= 0 {
		p.MinTeams = 1
	}
	if p.MaxTeams < p.MinTeams {
		p.MaxTeams = p.MinTeams
	}
	if p.MinBidPoints <= 0 {
		p.MinBidPoints = 1
	}
	if p.MaxBidPoints < p.MinBidPoints {
		p.MaxBidPoints = p.MinBidPoints
	}
	return p
}

// initDPTables allocates and initializes the DP value tables and
// backtracking pointer arrays. All DP cells start at negative infinity
// except dpPrev[0][0] = 0 (zero budget, zero teams selected).
func initDPTables(nTeams, maxBudget, maxTeams int) *dpTables {
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

	return &dpTables{
		dpPrev:     dpPrev,
		dpCur:      dpCur,
		prevBudget: prevBudget,
		prevCount:  prevCount,
		chosenBid:  chosenBid,
		maxBudget:  maxBudget,
		maxTeams:   maxTeams,
	}
}

// solveDPTransitions runs the knapsack-style DP over all teams.
// For each team, it considers skip (bid=0) or bid in [minBid, maxBid],
// updating the optimal objective value at each (budget, teamCount) state.
func solveDPTransitions(dp *dpTables, teams []Team, params AllocationParams) {
	negInf := math.Inf(-1)

	options := make([]int, 0, (params.MaxBidPoints-params.MinBidPoints+1)+1)
	options = append(options, 0)
	for bid := params.MinBidPoints; bid <= params.MaxBidPoints; bid++ {
		options = append(options, bid)
	}

	for i := 0; i < len(teams); i++ {
		for b := 0; b <= dp.maxBudget; b++ {
			for k := 0; k <= dp.maxTeams; k++ {
				dp.dpCur[b][k] = negInf
			}
		}

		t := teams[i]
		for b := 0; b <= dp.maxBudget; b++ {
			for k := 0; k <= dp.maxTeams; k++ {
				base := dp.dpPrev[b][k]
				if math.IsInf(base, -1) {
					continue
				}
				for _, bid := range options {
					if b+bid > dp.maxBudget {
						continue
					}
					k2 := k
					if bid > 0 {
						k2 = k + 1
					}
					if k2 > dp.maxTeams {
						continue
					}

					v := base + objective(t, bid)
					cur := dp.dpCur[b+bid][k2]
					if v > cur {
						dp.dpCur[b+bid][k2] = v
						dp.prevBudget[i+1][b+bid][k2] = b
						dp.prevCount[i+1][b+bid][k2] = k
						dp.chosenBid[i+1][b+bid][k2] = bid
						continue
					}
					if v == cur {
						prevBid := dp.chosenBid[i+1][b+bid][k2]
						if prevBid < 0 || bid < prevBid {
							dp.prevBudget[i+1][b+bid][k2] = b
							dp.prevCount[i+1][b+bid][k2] = k
							dp.chosenBid[i+1][b+bid][k2] = bid
						}
					}
				}
			}
		}

		dp.dpPrev, dp.dpCur = dp.dpCur, dp.dpPrev
	}
}

// selectBestSolution scans the final DP table to find the state with the
// highest objective value. Ties are broken by preferring higher budget
// spend, then fewer teams.
func selectBestSolution(dp *dpTables, params AllocationParams) (bestBudget, bestTeams int, found bool) {
	negInf := math.Inf(-1)
	bestV := negInf
	bestBudget = -1
	bestTeams = -1

	for b := 0; b <= dp.maxBudget; b++ {
		for k := params.MinTeams; k <= dp.maxTeams; k++ {
			v := dp.dpPrev[b][k]
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
		return 0, 0, false
	}
	return bestBudget, bestTeams, true
}

// backtrackBids walks the backtracking pointers from the best solution
// state back through each team decision to reconstruct the bid map.
func backtrackBids(dp *dpTables, teams []Team, bestBudget, bestTeams int) map[string]int {
	bids := make(map[string]int)
	b := bestBudget
	k := bestTeams
	for i := len(teams); i >= 1; i-- {
		bid := dp.chosenBid[i][b][k]
		if bid < 0 {
			break
		}
		if bid > 0 {
			bids[teams[i-1].ID] = bid
		}
		pb := dp.prevBudget[i][b][k]
		pk := dp.prevCount[i][b][k]
		if pb < 0 || pk < 0 {
			break
		}
		b = pb
		k = pk
	}
	return bids
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
