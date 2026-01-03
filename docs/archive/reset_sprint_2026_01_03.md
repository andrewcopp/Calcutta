# Reset Sprint — Core/Derived + Run Viewer + Go-First (2026-01-03)

## Objective
Re-align the ML/analytics pipeline around a small set of clear concepts:

- **Returns**: predict team success.
- **Investments**: predict what the auction market will pay.
- **Entries**: choose an optimal portfolio (bids) using returns + investments.
- **Evaluation**: backtest/score entries across historical years.

Frontend is **read-only**: browse runs created by CLI/Airflow.

## Non-goals
- No “run from the frontend” registry.
- No multi-schema sprawl beyond `core` and `derived`.
- No duplicate evaluation implementations in Python and Go.

## Decisions (locked)
-[x] **Schemas**: use `core` for source-of-truth entities and `derived` for rebuildable artifacts.
-[x] **Run identity**: `id UUID` as primary identifier for run-like tables.
-[x] **Human readable name**: `name TEXT` on run-like tables.
-[x] **Units**:
  - In-game: `*_points` (budget, bids, team scoring).
  - Real money: `*_cents` (payout amounts).
-[x] **Bids column naming**: canonical `bid_points`.
-[x] **Market prediction output**: store `predicted_share_of_pool` only; derive `predicted_points` at read time via `n_entries * budget_points`.
-[x] **Language boundary**:
  - Go owns simulation + evaluation + aggregation.
  - Python exists only for training/inference when it must use sklearn-style tooling.

## Target derived simulation model (naming + grouping)
We need a first-class grouping object for “these N simulations belong together”.

### `derived.simulation_states`
Represents the starting tournament state (wins/byes/eliminated per team).

- `id UUID PK`
- `tournament_id UUID FK -> core.tournaments(id)`
- `name TEXT NOT NULL`
- `source TEXT NOT NULL`
- `description TEXT NULL`
- `created_at`, `updated_at`, `deleted_at`

### `derived.simulation_state_teams`
Per-team state rows for a `simulation_state`.

- `id UUID PK`
- `simulation_state_id UUID FK -> derived.simulation_states(id)`
- `team_id UUID FK -> core.teams(id)`
- `wins INT NOT NULL`
- `byes INT NOT NULL`
- `eliminated BOOL NOT NULL`
- audit columns

### `derived.simulated_tournaments`
**Run header**. This is the grouping object for N Monte Carlo simulations.

- `id UUID PK`
- `tournament_id UUID FK -> core.tournaments(id)`
- `simulation_state_id UUID FK -> derived.simulation_states(id)`
- `name TEXT NOT NULL`
- `n_sims INT NOT NULL`
- `seed INT NOT NULL`
- `probability_source_key TEXT NOT NULL`
- `debug_mode BOOL NOT NULL DEFAULT FALSE`
- audit columns

### `derived.simulated_teams`
Team outcomes inside a `simulated_tournament` for each simulation number.

- `id UUID PK`
- `simulated_tournament_id UUID FK -> derived.simulated_tournaments(id)`
- `sim_number INT NOT NULL`
- `team_id UUID FK -> core.teams(id)`
- `wins INT NOT NULL`
- `byes INT NOT NULL`
- `eliminated BOOL NOT NULL`
- audit columns

## Target derived evaluation outputs
Evaluation should store:

- **Per-simulation, per-entry outcomes** (optional / debug, potentially large).
- **Per-run aggregated entry performance** (required).

### Payout storage and normalization
- Store money payouts as `payout_cents`.
- Define `normalized_payout = payout_cents / max_payout_cents_for_that_sim`.
- Aggregate as `mean_normalized_payout`, `median_normalized_payout`, `p_top1`, `p_in_money`.

## Pipeline run viewer (frontend)
The frontend is a browse UI:

- Returns: list runs, click run, click year, show results.
- Investments: list predicted market share runs, click year, show predicted shares.
- Entries: list strategy generation runs (returns + investment + optimizer), click year, show:
  - mean_normalized_payout
  - p_top1
  - p_in_money
  - our bids

## Retention policy (simulations)
Default should be “derived but ephemeral”.

-[ ] Implement default retention: keep only aggregated evaluation summaries.
-[ ] Implement debug retention: keep either sampled simulations or full simulations for the newest N runs.

## Migration plan (high-level)
We will do a migration-heavy upfront change to prevent ongoing confusion.

### 1) Move + rename simulation lineage tables into `derived`
From current tables (existing names):
- `analytics.tournament_state_snapshots` -> `derived.simulation_states`
- `analytics.tournament_state_snapshot_teams` -> `derived.simulation_state_teams`
- `analytics.tournament_simulation_batches` -> `derived.simulated_tournaments`
- `analytics.simulated_tournaments` -> `derived.simulated_teams`

Update all foreign keys that reference these tables.

### 2) Rename payout columns for unit correctness
- `payout_points` -> `payout_cents`
- Ensure queries/aggregations use `max_payout_cents` aliasing where appropriate.
- Standardize aggregated fields to `mean_normalized_payout`, `median_normalized_payout`.

### 3) Standardize bids
- Standardize `bid_points` everywhere for recommended bids.

### 4) Remove Python evaluation paths
- Keep Python only for training/inference.
- Ensure Airflow/DAG uses Go tasks for simulation + evaluation.

## Checklist
### Documentation / alignment
-[x] Record sprint decisions + target schema

### Core/Derived schema + migrations
-[ ] Create migration: move/rename simulation tables into `derived` and rename to `simulation_*` / `simulated_*`
-[ ] Create migration: rename payout columns `payout_points` -> `payout_cents`
-[ ] Create migration: standardize bids to `bid_points`
-[ ] Create migration: add `name` to run-like tables where needed

### Go (authoritative implementation)
-[ ] Update Go tournament simulation writer to write `derived.simulated_tournaments` + `derived.simulated_teams`
-[ ] Update Go simulated calcutta evaluator to read from `derived.simulated_*` tables
-[ ] Update Go evaluator to write payouts as `payout_cents`
-[ ] Update sqlc queries and regenerate code

### Python (deprecations)
-[ ] Remove/disable Python evaluation paths
-[ ] Keep only training/inference + writing predictions needed for Investments

### Frontend
-[ ] Replace/extend Simulations page to browse runs (Returns/Investments/Entries)
-[ ] Add drill-down: run -> year -> results

## Open questions
-[ ] Confirm the final names for evaluation tables in `derived` (whether to keep `entry_simulation_outcomes`/`entry_performance` naming or rename to `derived.simulated_entry_outcomes`/`derived.entry_performance`)
-[ ] Confirm where predicted market share run headers live and what their minimal schema is
