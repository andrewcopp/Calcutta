package main

import (
	"fmt"
	"math"
)

func dpAllocateBids(predPoints []float64, baseBids []float64, budget int, minTeams int, maxTeams int, minBid int, maxBid int) ([]int, int, error) {
	n := len(predPoints)
	if n != len(baseBids) {
		return nil, 0, fmt.Errorf("predPoints and baseBids length mismatch")
	}
	negInf := math.Inf(-1)

	dp := make([][]float64, maxTeams+1)
	for k := 0; k <= maxTeams; k++ {
		dp[k] = make([]float64, budget+1)
		for b := 0; b <= budget; b++ {
			dp[k][b] = negInf
		}
	}
	dp[0][0] = 0

	type parent struct {
		prevK  int
		prevB  int
		choice int
	}
	parents := make([][][]parent, n+1)
	for i := 0; i <= n; i++ {
		parents[i] = make([][]parent, maxTeams+1)
		for k := 0; k <= maxTeams; k++ {
			parents[i][k] = make([]parent, budget+1)
		}
	}

	for i := 0; i < n; i++ {
		newDP := make([][]float64, maxTeams+1)
		for k := 0; k <= maxTeams; k++ {
			newDP[k] = make([]float64, budget+1)
			for b := 0; b <= budget; b++ {
				newDP[k][b] = negInf
			}
		}

		for k := 0; k <= maxTeams; k++ {
			for b := 0; b <= budget; b++ {
				cur := dp[k][b]
				if cur == negInf {
					continue
				}

				if cur > newDP[k][b] {
					newDP[k][b] = cur
					parents[i+1][k][b] = parent{prevK: k, prevB: b, choice: 0}
				}

				if k == maxTeams {
					continue
				}
				for x := minBid; x <= maxBid; x++ {
					if b+x > budget {
						break
					}
					den := baseBids[i] + float64(x)
					val := 0.0
					if den > 0 {
						val = predPoints[i] * float64(x) / den
					}
					next := cur + val
					if next > newDP[k+1][b+x] {
						newDP[k+1][b+x] = next
						parents[i+1][k+1][b+x] = parent{prevK: k, prevB: b, choice: x}
					}
				}
			}
		}

		dp = newDP
	}

	bestK := -1
	bestVal := negInf
	for k := minTeams; k <= maxTeams; k++ {
		v := dp[k][budget]
		if v > bestVal {
			bestVal = v
			bestK = k
		}
	}
	if bestK == -1 || bestVal == negInf {
		return nil, 0, fmt.Errorf("no feasible allocation found")
	}

	bids := make([]int, n)
	k := bestK
	b := budget
	for i := n; i >= 1; i-- {
		p := parents[i][k][b]
		bids[i-1] = p.choice
		k = p.prevK
		b = p.prevB
	}

	selected := 0
	for _, x := range bids {
		if x > 0 {
			selected++
		}
	}

	return bids, selected, nil
}
