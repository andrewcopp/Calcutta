# Data Science Plans (from Go Sandbox)

This file captures the most useful ideas and workflow concepts from the retired Go-based sandbox so they can be continued in the Python-based `data-science/` area.

## What the Go sandbox was doing

- A small CLI designed for repeatable analysis against the Calcutta dataset.
- It centered on the separation of:
  - returns prediction (expected tournament points / points share)
  - investment prediction (market bid share)
  - ROI / value signals derived deterministically from the above

## Concepts worth keeping

### 1) Treat the DB / snapshot as source of truth

- Training and backtests should come from the authoritative dataset (historically the Postgres DB).
- Exporting derived datasets (CSV/Parquet) is fine, but treat them as build artifacts.
- Python direction: use the analytics snapshot export (zip -> parquet) as a reproducible “frozen” dataset input.

### 2) Leakage control: exclude your own entry from “market” features

- When computing market investment totals/shares, support excluding a specific entry name (the Go sandbox used `-exclude-entry-name`), to:
  - reduce strategy leakage when evaluating investment models
  - measure “cannibalization” (how your bids change ownership / pricing)
- Python direction: bake this into dataset derivations (i.e., compute both “market” and “market excluding X”).

### 3) Always keep a simple baseline yardstick

- Maintain a seed-only baseline model (both for returns and for investment) as the minimum bar.
- Support rolling training windows by year (e.g. “use last N tournament years” vs “all history”).
- Evaluation should be by year splits (train on past, test on holdout year).

### 4) Optimize bids as a constrained portfolio problem (not greedy)

The Go sandbox implemented an integer dynamic programming optimizer (“knapsack style”) that maximized expected entry points under constraints.

- **Ownership model** (given base market bid `B_i` and our bid `x_i`):
  - `own_i(x_i) = x_i / (B_i + x_i)`
- **Expected contribution**:
  - `EV_i(x_i) = pred_points_i * own_i(x_i)`
- **Objective**:
  - maximize `sum_i EV_i(x_i)`
- **Constraints**:
  - total budget (default 100)
  - bid on between 3 and 10 teams
  - bid per team between 1 and 50 dollars

Python direction: re-implement this optimizer (same objective/constraints) so that portfolio selection is “global optimal” under the team-count cap.

### 5) Prefer probability/EV framing and track uncertainty

- Treat realized outcomes as one draw; avoid “chasing” cinderella runs.
- Favor probability-based or EV-based models; include uncertainty where possible.
- Practical early approach:
  - keep deterministic EV predictions
  - estimate uncertainty via residual bootstraps / historical dispersion

### 6) Investment modeling idea: shrink noisy “within-seed” signals toward seed means

A useful pattern in the Go sandbox was “shrinkage” toward a stable seed baseline:

- Start with seed mean bid share.
- Add a KenPom-based (or other strength proxy) correction.
- Shrink that correction toward 0 with a tuned alpha.
- Tune alphas using leave-one-calcutta-out cross-validation.
- Optionally tune different alphas for different seed segments (top 4, top 8, rest).

This is a good bias/variance tradeoff for small sample sizes.

## Research ideas worth continuing (condensed)

These are the strongest hypotheses from the Go sandbox backlog to consider pursuing in Python.

- **Within-seed pod mispricing (“Overlooked #3”)**
  - within a 4-team pod, the “3rd perceived best” may be systematically underbid
  - needs a within-pod strength ranking signal
- **Region strength effects (“Region of Death”)**
  - perceived path difficulty may discount bids on teams in a brutal region
- **Top-heavy payout bias**
  - market may overpay for championship equity vs expected points
  - suggests modeling late-round equity (Final Four / title probability)
- **Brand tax / familiarity premium**
  - household-name programs may be persistently overbid beyond what strength explains

## Concrete next steps in `data-science/`

### A) Dataset standardization

- Input: analytics snapshot zip (downloaded or local) -> parquet-per-table.
- Derive canonical tables suitable for modeling:
  - per-team table with team metadata + realized points + market bid share
  - versions excluding a named entry

### B) Rebuild the “harness” loops

- “Backtest harness”: iterate years, train on past years, test on holdout year.
- “Report harness”: produce a human-readable markdown report per year:
  - predicted returns (and shares)
  - predicted market bids (and shares)
  - optimized portfolio + expected ROI
  - realized performance vs field

### C) Rebuild the portfolio optimizer

- Implement the DP optimizer with the same constraints.
- Add hooks to compare:
  - greedy allocation vs DP optimal allocation
  - sensitivity to budget/team-count caps

### D) Start with these model baselines

- **Returns**: seed-only expected points, then “strength proxy” (KenPom if present in snapshot/DB).
- **Investment**: seed-only, then shrinkage toward seed mean using strength proxy.

## Notes

- The old Go sandbox also defined a set of investment model variants and evaluation metrics. If/when you want to replicate those in Python, use the same north-star metric concept:
  - “how well do we predict within-seed residuals, normalized by historical within-seed dispersion?”.
