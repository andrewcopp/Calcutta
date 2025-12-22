# Data Science Sandbox (Predict Returns, Predict Investment, Compute ROI)

## Goal
Build a separate, iteration-friendly “data science sandbox” that:

- Predicts **team returns** (expected tournament points / expected value under Calcutta rules)
- Predicts **community investment** (expected total bids per team, pre-auction)
- Computes **predicted ROI** so we can allocate our own portfolio by **marginal ROI**

This is explicitly intended to support *pre-auction* strategy in a **blind simultaneous auction**.

## Non-goals (for now)
- Not building a polished web UI.
- Not embedding experimental models directly into the production API.
- Not using historical CSVs as the source of truth (they are not perfectly clean).

## Source of truth
- **Postgres DB** is the authoritative dataset.
- Use the DB to generate “training” and “backtest” datasets.
- Exporting derived datasets (CSV/Parquet) for analysis is OK, but they should be treated as **build artifacts**.

## Why two models?
In this Calcutta format, “value” is not just picking winners.

- A team can be great and still be a bad investment if it is **over-subscribed**.
- A team can be mediocre and still be a great investment if it is **under-subscribed**.

So the sandbox centers on:

1. **Returns model**: how many points a team is expected to generate
2. **Investment model**: how much total community investment a team is expected to attract
3. **ROI model**: a deterministic combination of (1) and (2)

## Key definitions
- **Returns**: expected points earned by the team under the scoring rules.
- **Investment**: total dollars bid on the team by the whole community.
- **Ownership share**: `my_bid / total_investment`
- **My expected points from a team**: `ownership_share * expected_team_points`
- **Team-level ROI (points/$)**: `expected_team_points / predicted_total_investment`

For comparing across seeds/years, we often prefer a **normalized ROI**:

- **Team points share**: `expected_team_points / total_expected_points_in_calcutta`
- **Team investment share**: `predicted_total_investment / total_predicted_investment_in_calcutta`
- **Normalized ROI**: `points_share / investment_share`

Interpretation:
- `Normalized ROI > 1.0` implies under-invested (a bargain)
- `Normalized ROI < 1.0` implies over-invested

## Sandbox MVP output
Given a target year (e.g. 2025), produce a table (and plots) at the team level:

- tournament year
- team, seed, region
- KenPom features (as available)
- `predicted_team_points`
- `predicted_total_investment`
- `predicted_normalized_roi`
- plus diagnostics (confidence intervals / prediction intervals where possible)

## Returns model: staged roadmap
### Baseline (seed-only)
- Estimate `P(reach round k | seed)` from historical tournaments.
- Convert to expected points using the scoring schedule.

### Next layer (KenPom)
- Add KenPom metrics as covariates.
- Prefer outputs as **round-advancement probabilities** (calibrated) rather than a single point estimate.

### Advanced
- Game-by-game simulation driven by team strength (KenPom AdjEM / offense/defense).

## Investment model: staged roadmap
Investment prediction is harder because it includes behavioral feedback.

### Baseline (seed-only)
- Predict expected investment per seed from historical Calcuttas.

### Within-seed pod dispersion
Model “within seed pod” differences:
- Example: among the four 1-seeds, which gets overbid/underbid?
- Candidate features:
  - KenPom rank / AdjEM rank within pod
  - brand proxies (historical appearances, historical over/under-valuation)
  - year-over-year recency signals (prior year seed-tier over/under investment)

### Seasonality / feedback
- Include features like “prior-year mean ROI for this seed” and “prior-year mean investment share for this seed”.
- Expect limited sample size; treat this as exploratory and likely require regularization.

### Output requirement
For game theory and marginal allocation, the investment model should ideally output:
- `E[total_investment]` and also a **distribution** (or at least variance)

Practical MVP approach:
- Predict mean investment and approximate uncertainty via historical residuals, bootstrapping, or a simple hierarchical model.

## Portfolio optimization (later, but design for it now)
Ultimately, you allocate bids to maximize **expected payout** (not just expected points).
That likely requires:

- Modeling the distribution of your portfolio points
- Modeling the distribution of *other entries’* portfolios via predicted investments
- Applying the payout table (top 5/7 splits) to simulated leaderboard outcomes

MVP will focus on per-team value signals and marginal ROI curves.

## Backtesting & evaluation
### Returns model evaluation
Preferred metrics depend on output type:
- If predicting advancement probabilities:
  - log loss / Brier score per round
  - calibration curves
- If predicting expected points:
  - MAE/RMSE on realized team points

### Bad beats, over-performance, and not overfitting
Some of the most memorable historical "best investments" are driven by low-probability outcomes:

- A heavily owned favorite can become a terrible realized pick due to a single upset.
- A lightly owned longshot can look like genius in hindsight due to an unlikely Cinderella run.

For modeling, this means:

- We should optimize for **expected value**, not "what happened last time".
- We should treat realized points and realized ROI as **one draw** from a distribution.
- We should prefer probability-based returns models (round advancement probabilities) so we can separate:
  - a good bet that lost (bad beat)
  - from a bad bet that won (lucky)

Practical implications for the sandbox:

- Report both **EV** and an **uncertainty measure** (prediction interval / variance proxy).
- Use backtests that score predictions by calibration (log loss / Brier), not just by picking the single best historical ROI.
- Use historical extremes (Cinderellas, crash-outs) to sanity-check ceilings/floors, but avoid training directly to "chase" those outcomes.

### Investment model evaluation
- MAE/RMSE on team total investment
- rank correlation within seed pods (did we correctly order “hot” vs “cold” teams?)
- calibration of prediction intervals (e.g., 80% PI contains truth ~80% of the time)

### Time-split strategy
- Cross-validate by year (train on past years, test on a held-out year).
- Be careful with “seasonality” features; they reduce usable training years.

## Visualizations (recommended)
To understand dispersion and “ugly duckling” effects:

- **Box plot / violin plot** of investment within each seed (or seed pod)
- **Scatter**: predicted returns vs predicted investment (quadrants)
- **Within-pod rank plot**: x = KenPom pod rank, y = investment share
- **Year-over-year trend**: seed investment share vs seed ROI

Note: Recharts doesn’t have a built-in box plot; for sandbox analysis it’s usually easiest to use Python plotting (seaborn/matplotlib) and export images.

## Existing related docs
- `docs/investability_theories.md` (contains “Ugly Duckling Theory”, stddev analysis ideas, and future analytics roadmap)
- `docs/analytics_future_enhancements.md` (seed normalization + bias score ideas)

This doc focuses on the *sandbox workflow + model architecture*.
