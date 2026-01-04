# Testing Strategies + In-Game Predictions: DAG Plan

## Context
We have two related DAGs:

### DAG 1: Testing Strategies
Goal: iterate on strategy components and evaluate them against historical calcuttas.

Steps:
- Predicted Returns
- Predicted Investments (market)
- Optimized Entry
- Synthetic Calcutta
- Simulated Tournaments / Calcutta Evaluation

### DAG 2: In-Game Predictions
Goal: generate matchup-level predictions used by tournament simulation + calcutta evaluation.

Steps:
- Predicted Game Outcomes
- Simulated Tournaments / Calcutta Evaluation

## Key architectural stance
Python runs ahead-of-time to generate predictions and registers both:
- the artifacts (row-level outputs)
- the run metadata (how they were produced)

Go owns:
- the MINLP allocator / portfolio construction
- orchestration and evaluation services

Frontend should:
- discover registered algorithms + runs
- allow selecting returns + market algorithms
- generate and save test entries
- trigger and inspect evaluation runs

## Scoping decisions (v1)
- Predicted game outcomes are the primitive (matchup-level).
  - Stored as discrete matchup probabilities so we can compose them into simulations and support forecasting from any tournament state.
- Predicted market share is scoped to a calcutta.
  - The calcutta implies the tournament (which teams to predict).
  - Training scope is the calcutta group (nullable). If null, train on all calcuttas.
- Calcutta scoring is applied later.
  - We can infer predicted points/returns by combining predicted game outcomes with the calcutta scoring rules.

## Evaluation modes

### A) Evaluate a Calcutta as-is
Goal: forecast final standings for a single calcutta from the current tournament state.

Notes:
- Realized outcomes can be computed directly from results without evaluation.
- Evaluations are used to compare “what we thought would happen” vs “what did happen” (who got lucky).

### B) Evaluate a single Calcutta entry for a single year
Goal: atomic capability for evaluation (a single entry in a single calcutta).

Note: this may not be exposed as a first-class UI workflow initially, but it is the foundational unit for benchmarking suites.

### C) Benchmark a group of entries across years
Goal: create a group of entries across multiple years using the same selected algorithms and compare performance across all years.

Context: the purpose of C is to avoid overfitting. If we only test an algorithm against one year, we can optimize for quirks of that year. We want to select algorithms that maximize our chance of success across years.

## Work plan (`- [ ]`)

### Immediate unblock / correctness
- [ ] Commit predicted-investment fallback fix (analytics_calcutta_predictions.sql + sqlc output)
- [ ] Implement pure `predicted_game_outcomes` -> predicted returns / expected value (no Monte Carlo)
  - [ ] Use bracket DP over matchup probabilities to compute per-team round advancement + expected points
  - [ ] Use this as the baseline for naive `predicted_market_share` (until market model is wired)
- [ ] Verify frontend smoke run UX end-to-end
  - [ ] /runs/2025 lists run
  - [ ] Returns page loads and is non-empty
  - [ ] Investments page loads and is non-empty
  - [ ] “Our entry” behavior matches optimizer output

### Testing Strategies DAG (registry + artifacts)
- [ ] Design schema v1
  - [ ] Calcutta-scoped market share
  - [ ] Tournament-scoped predicted game outcomes (matchup-level)
  - [ ] Minimal run metadata tables with `params_json` to avoid schema lock-in
  - [ ] Algorithm metadata + per-tournament runs
    - [ ] High-level algorithm tables describe the algorithm itself (stable identity)
    - [ ] Child run tables record outputs for each tournament/season the algorithm was applied to

- [ ] Add algorithm registry + run metadata tables (derived.*)
  - [ ] `derived.algorithms` (generic registry by kind)
  - [ ] Predicted game outcomes: algorithm metadata + per-tournament runs + matchup artifact table
  - [ ] Predicted market share: algorithm metadata + per-calcutta runs + market share artifact table

- [ ] Update writers/readers
  - [ ] Python writers: write run metadata first, then artifacts linked to run
  - [ ] Migrate `run_ridge_regression.py` / market-share DB writer off `lab_bronze.*` and onto current `core/derived`
    - [ ] Write a `derived.market_share_runs` row and link artifacts via `derived.predicted_market_share.run_id`
    - [ ] Ensure it can write both calcutta-scoped and (legacy) tournament-scoped rows as needed
  - [ ] Go readers/services:
    - [ ] Accept explicit run_id selection
    - [ ] Game outcomes: default to latest for the tournament
    - [ ] Market share: default to latest for the calcutta if it exists; otherwise error (no naive fallback)

- [ ] Discovery endpoints
  - [ ] list algorithms by kind
  - [ ] list runs for (algorithm, tournament) and (algorithm, calcutta)

- [ ] Strategy generation workflow
  - [ ] Select game-outcomes run + market share run + optimizer key
  - [ ] Generate and save optimized entry (persist provenance)
  - [ ] Record per-team diagnostics (original ROI vs adjusted ROI)

- [ ] Evaluation workflow
  - [ ] Configure evaluation (exclude/include calcuttas, n_sims, seed)
  - [ ] Trigger evaluation run and browse results
  - [ ] Headline metrics
    - [ ] Mean normalized payout (primary)
    - [ ] P(Top 1)
    - [ ] P(In Money)
    - [ ] Expected points vs actual points
    - [ ] Actual finish position

- [ ] Evaluation modes
  - [ ] A) Evaluate a Calcutta as-is (in-app predicted finishes)
  - [ ] B) Evaluate a single entry for a single year (atomic; UI may be deferred)
  - [ ] C) Benchmark a group of entries across years (same algorithms across years)

### Naming glossary
- Algorithm: stable definition (e.g. KenPom v1, Ridge Regression v2)
- Run: execution/application of an algorithm that produces artifacts
- Artifact: row-level outputs (matchup probabilities, per-team market share, etc.)
- Suite: cross-year recipe (algorithm choices + optimizer key + evaluation settings)
- SuiteCalcuttaEvaluation: one execution of a suite for a specific calcutta (which implies tournament/year)
- StrategyGenerationRun: persisted optimized entry (existing)

### In-Game Predictions DAG
- [ ] Schema v1 for `predicted_game_outcomes`
  - [ ] run metadata table
  - [ ] artifact table linked to run
- [ ] Ensure tournament simulation selects a specific game-outcomes run

### Cleanup
- [ ] Identify and retire exploratory/legacy tables/endpoints/UI once registry model is wired
