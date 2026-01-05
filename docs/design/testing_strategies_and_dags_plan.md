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

## Frontend IA (v1): Lab + Sandbox

### Navigation
- Replace: Simulations / Predictions / Evaluations / Runs
- With: Lab / Sandbox

### Lab

#### Tab: Advancements
Source of truth:
- `derived.game_outcome_runs`
- `derived.predicted_game_outcomes`

Behavior:
- List registered advancement algorithms (game-outcomes algorithms)
- Click algorithm -> show algorithm metadata + list of tournaments it has runs for
- Click tournament -> show per-team advancement probabilities

Per-team advancement table:
- Show cumulative probability of reaching each round (no points / returns)
- Default sort: highest championship probability
- Sanity check expectation: totals should be ~100%, 200%, 400%, ... by round (tournament structure dependent)

#### Tab: Investments
Source of truth:
- `derived.market_share_runs`
- `derived.predicted_market_share`

Behavior:
- List registered investment algorithms (market-share algorithms)
- Click algorithm -> show algorithm metadata (model + features used) + list of calcuttas it has runs for
- Click calcutta -> show run metadata (training data used) + per-team predicted market share table

Per-team investments table:
- Show `predicted_market_share` as a percentage (expect very small values; allow 3-4 decimal places)
- Show `rational_market_share` as the baseline implied by predicted returns (expected points under calcutta scoring) under a naive equal-ROI market assumption (not multiplied by total pool)
- Show `delta_percent` as percent difference derived from the ratio between predicted and rational market share (handle `rational=0` explicitly)
- Color semantics: negative = green (buying opportunity), positive = red (over-invested)

### Sandbox
Source of truth:
- TestSuites and their executions against all calcuttas (Mode C)

Behavior:
- List TestSuites that have been run
- Each suite name describes the advancement algorithm, investment algorithm, and entry optimizer
- Click suite -> show suite metadata (algorithms used, simulations per test, excluded entry name) + list of per-calcutta results
- Click a per-calcutta result -> show the entry that produced those stats

Displayed per-calcutta headline stats:
- Expected Position (Calcutta ranking)
- Mean Normalized Payout
- P(Top 1)
- P(In Money)
- Realized Finish (deterministic finish vs historical results)

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
- [x] Only then: run Monte Carlo evaluation for distributions
  - [x] Suite calcutta evaluation smoke test passed (end-to-end: enqueue -> running -> succeeded -> results)

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
    - [x] Market share: default to latest for the calcutta if it exists; otherwise error (no naive fallback)

- [x] Discovery endpoints
  - [x] list algorithms by kind (`GET /api/analytics/algorithms?kind=...`)
  - [x] list game-outcome runs (`GET /api/analytics/tournaments/{id}/game-outcome-runs`)
  - [x] list market-share runs (`GET /api/analytics/calcuttas/{id}/market-share-runs`)
  - [x] latest run IDs for calcutta (`GET /api/analytics/calcuttas/{id}/latest-prediction-runs`)

- [ ] Additional endpoints for Lab + Sandbox browsing
  - [x] Tournament advancement view (tournament-scoped; no points)
    - [x] `GET /api/analytics/tournaments/{id}/predicted-advancement?game_outcome_run_id=...`
  - [x] Calcutta market share view (share-based; not pool points)
    - [x] `GET /api/analytics/calcuttas/{id}/predicted-market-share?market_share_run_id=...&game_outcome_run_id=...`
  - [ ] Suite results browsing
    - [x] List suites
    - [x] Suite detail (metadata + per-calcutta results)
    - [x] Result detail (generated entry + provenance)
  - [ ] Cleanup note: these Lab endpoints are additive; identify and retire legacy predicted returns/investment endpoints and pages once the new UI is wired

- [ ] Sandbox persistence (inputs + outputs only; no simulation samples)
  - [ ] Persist suite definition (advancement algorithm/run selector, investment algorithm/run selector, optimizer, evaluation settings)
  - [ ] Persist suite execution metadata (created_at, seed, n_sims, excluded entry name)
  - [ ] Persist per-calcutta headline results needed by UI (expected position, mean normalized payout, P(top 1), P(in money), realized finish)
  - [ ] Persist a reference to the generated entry/provenance for drill-down

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
  - [x] Suite calcutta evaluation workflow (single calcutta)
    - [x] Select `game_outcome_run_id` + `market_share_run_id` + `optimizer_key`
    - [x] Persist provenance in `derived.suite_calcutta_evaluations` (including `strategy_generation_run_id` + `calcutta_evaluation_run_id`)
    - [x] API + worker execute evaluation and expose results via `calcutta_evaluation_run_id`

- [ ] Evaluation modes
  - [ ] A) Evaluate a Calcutta as-is (in-app predicted finishes)
  - [ ] B) Evaluate a single entry for a single year (atomic; UI may be deferred)
  - [ ] C) Benchmark a group of entries across years (same algorithms across years)

- [ ] Frontend: Lab + Sandbox UI
  - [x] Replace nav with Lab / Sandbox and remove legacy pages from navigation
  - [ ] Coverage: use UI to identify missing runs (algorithms not yet run for a tournament / calcutta)
  - [ ] Lab: Advancements tab
    - [x] Algorithm list (kind = game outcomes)
    - [x] Algorithm detail -> tournaments with runs
    - [x] Tournament detail -> per-team advancement probabilities (sortable; default by championship probability)
  - [ ] Lab: Investments tab
    - [x] Algorithm list (kind = market share)
    - [x] Algorithm detail -> calcuttas with runs
    - [x] Calcutta detail -> per-team market share with rational + predicted + delta_percent
  - [ ] Sandbox: Suites browser
    - [x] List suites
    - [x] Suite detail -> per-calcutta results list + suite metadata
    - [x] Result detail -> show generated entry and its provenance (dedicated detail route: `/sandbox/evaluations/:id`)

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
- [x] Ensure tournament simulation selects a specific game-outcomes run

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
