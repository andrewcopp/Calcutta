# Data Science Ideas Backlog

This file is a lightweight backlog of research ideas, distilled from the retired Go sandbox and related experiment notes.

## Investment (market bid share) ideas

- **Within-seed pod mispricing (“Overlooked #3”)**
  - Hypothesis: in a 4-team pod (especially 1-seeds), the “3rd perceived best” is systematically underbid.
  - Needs: within-pod strength proxy (KenPom rank/AdjEM, odds, etc.).

- **Region of Death / path narrative**
  - Hypothesis: teams in unusually strong regions receive an investment discount relative to similar-strength teams in easier regions.
  - Needs: region strength features (top-4/top-8 avg strength) + strength proxy.

- **Top-heavy payout bias (championship equity premium)**
  - Hypothesis: market overbids teams with high title equity relative to expected points.
  - Needs: late-round probability features (`p_final_four`, `p_title`) or proxies.

- **Brand tax / familiarity premium**
  - Hypothesis: household-name programs are persistently overbid after controlling for strength.
  - Needs: appearances / recency / brand score proxies.

- **Conference narratives**
  - Hypothesis: conference effects remain after controlling for strength (inflation/discount by conference).
  - Needs: conference data + controls.

- **“Bubble fraud” discount**
  - Hypothesis: disagreement between seed and external strength/record creates systematic under/overbidding.
  - Needs: record, SOS, or better strength features.

- **Winner’s curse / overpayment factor**
  - Hypothesis: sealed-bid dynamics create systematic overpayment gaps; could be used to adjust ROI when bidding.
  - Needs: per-bid distribution data (not just team totals).

- **“Upset chic” inflation**
  - Hypothesis: trendy upset seed bands (10–12) are overbid conditional on EV.
  - Needs: matchup/upset-probability features for clean testing.

## Modeling / evaluation ideas

- **Shrinkage to seed means**
  - Start with seed mean bid share; add a noisy signal (KenPom/returns proxy) and shrink toward seed mean.
  - Tune shrink alphas via year/Calcutta-held-out CV; consider different alphas for top 4 / top 8 / others.

- **Within-seed evaluation (“residual correctness”)**
  - Track metrics that emphasize ordering/variance within seeds/pods, not just global MAE.

## Portfolio / optimization ideas

- **DP (knapsack) portfolio optimizer**
  - Replace greedy marginal-ROI bidding with a global optimum under team-count caps and integer bid constraints.

- **Portfolio correlation premium**
  - Hypothesis: market prices some teams for diversification/correlation reasons (region/path complementarity).

