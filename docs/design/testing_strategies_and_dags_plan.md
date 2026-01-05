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

## Forecasting vs evaluation (important)

### Forecasting / predicted returns (no Monte Carlo)
We want a deterministic, repeatable path from matchup-level probabilities to per-team expected points.

Definition:
- Input: bracket structure + `derived.predicted_game_outcomes` (matchup probabilities)
- Output: per-team:
  - P(reach each round)
  - expected points in *Calcutta scoring units*

Implementation strategy:
- Use a bracket dynamic program (DP) over the tournament tree.
- This is NOT the same thing as calcutta evaluation.

Reasoning:
- Forecasting should not depend on Monte Carlo samples.
- We want predictable behavior and debuggable provenance (missing matchup probs should fail fast).
- “Returns” is a property of the game-outcomes run + scoring rules, not of an evaluation run.

### Evaluation / benchmarking (Monte Carlo)
Monte Carlo is reserved for:
- simulating full tournament outcome distributions (tail risk)
- evaluating a specific entry (payout distribution)
- benchmarking suites across years

Evaluation is allowed to be stochastic and expensive; it produces distributions and confidence bands.

## Expected-value semantics (wins + byes)
Important subtlety: our scoring function is defined as:

`core.calcutta_points_for_progress(calcutta_id, wins, byes)`

and it awards points for `win_index <= (wins + byes)`.

Implications:
- For non-First-Four teams, `byes=1` (they are already “in Round of 64”), so they start with progress=1.
- For First Four teams, `byes=0` and they must win the play-in to reach progress=1.
- Championship win corresponds to progress=7.

The DP evaluator must be aligned to these semantics; otherwise investments/returns can look “off” for play-in teams.

## Run-selection invariants

### Game outcomes (returns) selection
- Game-outcomes are tournament-scoped.
- Consumers should select a specific `derived.game_outcome_runs.id`.
- Default behavior: select latest run for tournament only for convenience; it should be explicit in suite executions.
- No legacy fallback: missing/mismatched matchup probabilities should fail fast.

### Market share (investment) selection
- Market share is calcutta-scoped.
- Consumers should select a specific `derived.market_share_runs.id`.
- Default behavior: if missing, error loudly (avoid silent fallback).
- If a developer wants local UX unblocked without running the market model, they must explicitly seed `derived.predicted_market_share` via a dev tool.

Status:
- `GET /api/analytics/calcuttas/{id}/predicted-investment` now accepts `market_share_run_id`.
- Default behavior is: select the latest `derived.market_share_runs` row for the calcutta and use `derived.predicted_market_share.run_id`.
- No legacy fallback: if no run exists, the endpoint should fail loudly.

### Legacy bridging
During migration:
- existing tables may still hold rows with `run_id IS NULL`.
- we must define when “legacy rows are acceptable” vs when to require run IDs.

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
- [x] Commit predicted-investment fallback fix (analytics_calcutta_predictions.sql + sqlc output)
- [x] Implement pure `predicted_game_outcomes` -> predicted returns / expected value (no Monte Carlo)
  - [x] Use bracket DP over matchup probabilities to compute per-team round advancement + expected points
  - [x] Provide an explicit dev-only seeding tool for `derived.predicted_market_share` (no runtime fallback)
- [x] Verify frontend smoke run UX end-to-end
  - [x] /runs/2025 lists run
  - [x] Returns page loads and is non-empty
  - [x] Investments page loads and is non-empty
  - [x] “Our entry” behavior matches optimizer output

### Practical smoke checklist (debug loop)
- [x] `derived.predicted_game_outcomes` has rows for the tournament
- [x] DP predicted returns are non-empty and stable (no Monte Carlo)
- [x] `derived.predicted_market_share` has rows for the tournament/calcutta (real model OR explicitly seeded baseline)
- [x] Investments endpoint returns non-zero values
- [x] Optimizer output matches “Our entry” behavior in UI
- [ ] Only then: run Monte Carlo evaluation for distributions

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
  - [x] Migrate `run_ridge_regression.py` / market-share DB writer off `lab_bronze.*` and onto current `core/derived`
    - [ ] Write a `derived.market_share_runs` row and link artifacts via `derived.predicted_market_share.run_id`
    - [x] Ensure it can write both calcutta-scoped and (legacy) tournament-scoped rows as needed
  - [ ] Go readers/services:
    - [x] Accept explicit run_id selection
    - [x] Game outcomes: default to latest for the tournament
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
- [x] Schema v1 for `predicted_game_outcomes`
  - [x] run metadata table (`derived.game_outcome_runs`)
  - [x] artifact table linked to run (`derived.predicted_game_outcomes.run_id`)
- [ ] Ensure tournament simulation selects a specific game-outcomes run

## Tooling notes (local)

### Generate predicted game outcomes (run-scoped)
This generator writes matchup probabilities to `derived.predicted_game_outcomes` with `run_id` populated and creates a `derived.game_outcome_runs` row.

Command:
- `go -C backend run ./cmd/generate-predicted-game-outcomes --season 2025`

### Seed a naive market baseline from PGO expected value (no Monte Carlo)
We added a local seeding tool that computes predicted expected value (DP over bracket) and writes a calcutta-scoped, run-linked baseline `derived.predicted_market_share`.

This is for local UX unblock only; it is NOT the market model.

This is not a runtime fallback: production workflows should fail if market data is missing.

Command:
- `go -C backend run ./cmd/seed-naive-market-share-from-pgo --calcutta-id <uuid>`
- `--dry-run` prints top teams and does not write

### Seed ridge-predicted market share from DB (no `out/` snapshots)
The ridge regression runner now trains/predicts from `core.*` and writes to `derived.predicted_market_share`.

Notes:
- It writes run-scoped rows (`run_id IS NOT NULL`) and creates a `derived.market_share_runs` row.
- This requires dropping obsolete unique indexes on `(tournament_id, team_id)`; local migration now handles this.

Migration:
- `20260105031000_drop_obsolete_predicted_market_share_indexes`

Command:
- `data-science/.venv/bin/python data-science/scripts/run_ridge_regression.py 2025`

### Calcutta evaluation
Monte Carlo evaluation should be run only after the DP returns + market baseline exist.

### Cleanup
- [ ] Identify and retire exploratory/legacy tables/endpoints/UI once registry model is wired
