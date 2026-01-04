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
- `predicted_market_share` is tournament-scoped for now.
  - Future: scope by “investment pool / calcutta group” (not implemented yet).
- Predicted returns are team-scoped and stored as round probabilities only.
  - Calcutta-scoring agnostic; scoring can be applied later.

## Evaluation modes

### A) Evaluate a Calcutta as-is
Goal: evaluate the actual recorded entries for a single calcutta using in-app predictions.

### B) Evaluate a single Calcutta entry for a single year
Goal: ad hoc experimentation by creating (or importing) a single entry, then evaluating it for a single tournament year.

### C) Benchmark a group of entries across years
Goal: create a group of entries across multiple years using the same selected algorithms and compare performance across all years.

Context: the purpose of C is to avoid overfitting. If we only test an algorithm against one year, we can optimize for quirks of that year. We want to select algorithms that maximize our chance of success across years.

## Work plan (`- [ ]`)

### Immediate unblock / correctness
- [ ] Commit predicted-investment fallback fix (analytics_calcutta_predictions.sql + sqlc output)
- [ ] Verify frontend smoke run UX end-to-end
  - [ ] /runs/2025 lists run
  - [ ] Returns page loads and is non-empty
  - [ ] Investments page loads and is non-empty
  - [ ] “Our entry” behavior matches optimizer output

### Testing Strategies DAG (registry + artifacts)
- [ ] Design schema v1
  - [ ] Tournament-scoped market share
  - [ ] Tournament-scoped team predicted returns (per-round probabilities)
  - [ ] Minimal run metadata tables with `params_json` to avoid schema lock-in
  - [ ] Algorithm metadata + per-tournament runs
    - [ ] High-level algorithm tables describe the algorithm itself (stable identity)
    - [ ] Child run tables record outputs for each tournament/season the algorithm was applied to

- [ ] Add algorithm registry + run metadata tables (derived.*)
  - [ ] `derived.algorithms` (generic registry by kind)
  - [ ] Predicted returns: algorithm metadata + per-tournament runs + team probability artifacts
  - [ ] Predicted market share: algorithm metadata + per-tournament runs + market share artifacts

- [ ] Update writers/readers
  - [ ] Python writers: write run metadata first, then artifacts linked to run
  - [ ] Go readers/services: accept explicit run_id selection, default to latest

- [ ] Discovery endpoints
  - [ ] list algorithms by kind
  - [ ] list runs for (algorithm, tournament)

- [ ] Strategy generation workflow
  - [ ] Select returns run + market share run + optimizer params
  - [ ] Generate and save entry (persist provenance)

- [ ] Evaluation workflow
  - [ ] Configure evaluation (exclude/include calcuttas, n_sims, seed)
  - [ ] Trigger evaluation run and browse results

- [ ] Evaluation modes
  - [ ] A) Evaluate a Calcutta as-is (in-app predicted finishes)
  - [ ] B) Evaluate a single entry for a single year (ad hoc)
  - [ ] C) Benchmark a group of entries across years (same algorithms across years)

### In-Game Predictions DAG
- [ ] Schema v1 for `predicted_game_outcomes`
  - [ ] run metadata table
  - [ ] artifact table linked to run
- [ ] Ensure tournament simulation selects a specific game-outcomes run

### Cleanup
- [ ] Identify and retire exploratory/legacy tables/endpoints/UI once registry model is wired
