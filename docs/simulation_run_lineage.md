# Simulation Run Lineage (Tournament Snapshots → Simulation Batches → Evaluation Runs)

## Goals
- Support **"as-of" evaluations** and a tournament rewind/time-machine.
- Provide immutable lineage so we can answer:
  - which tournament state was used?
  - which simulations were used?
  - which evaluation results came from them?
- Enable safe caching and deletion policies.

## Core concepts

### 1) Tournament state snapshots
A **tournament state snapshot** is the authoritative input describing which game results are locked.

Use cases:
- **Pre-tournament** leaderboard (0 locked games)
- After play-in games
- After each completed game (live leaderboard refresh)
- **Counterfactual** hall-of-fame (e.g., "as if game X never happened")
- End-state snapshot (all games locked)

Proposed data shape:
- `tournament_state_snapshots`
  - `id`
  - `tournament_id`
  - `created_at`
  - `source` (manual, ingest, derived)
  - `description`
- `tournament_state_snapshot_games`
  - `tournament_state_snapshot_id`
  - `game_id`
  - `winner_team_id`
  - `decided_at`

Invariant:
- A snapshot is immutable once created.

### 2) Simulation batches
A **simulation batch** is the output of running `n_sims` bracket simulations given:
- bracket structure
- win-probability source (model)
- a tournament state snapshot (locked results)
- random seed

Use cases:
- cache expensive simulation results
- enable comparisons across snapshots ("how odds changed")
- safely delete batches not referenced by any evaluation

Proposed data shape:
- `simulation_batches`
  - `id`
  - `tournament_id`
  - `tournament_state_snapshot_id`
  - `n_sims`
  - `seed`
  - `probability_source_key` (e.g., kenpom-v1, uploaded-model-xyz)
  - `created_at`

Output table:
- `simulated_tournaments`
  - `simulation_batch_id`
  - `sim_id`
  - `team_id`
  - `wins`, `byes`
  - (optional) `eliminated`

### 3) Evaluation runs
An **evaluation run** computes entry outcomes given:
- one simulation batch
- one calcutta snapshot (entry set + bids + payouts + scoring)

Use cases:
- compute predicted leaderboards and payout distributions
- track which results are shown in UI at a given time
- compare variant calcuttas (hot-swap an entry)

Proposed data shape:
- `evaluation_runs`
  - `id`
  - `simulation_batch_id`
  - `calcutta_snapshot_id`
  - `created_at`
  - `purpose` (live, hall_of_fame, sandbox_submission)

Outputs keyed by `evaluation_run_id`:
- `entry_simulation_outcomes`
- `entry_performance`

## Retention and deletion
- Simulation batches may be deleted only if:
  - no `evaluation_runs` reference them
- Evaluation outputs may be regenerated if their inputs still exist (snapshot + batch + calcutta snapshot)

## Caching policy (Airflow)
- Create a new tournament snapshot after each game is finalized.
- Periodically create a simulation batch for the latest snapshot (cadence: configurable).
- Create an evaluation run for the latest snapshot and the primary calcutta snapshot.

## Notes
- The core app should serve cached results.
- Users should not trigger simulation/evaluation on page load.
