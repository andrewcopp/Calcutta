# Codebase Partitions

The codebase is organized into three logical partitions: **Core**, **Sandbox**, and **Lab**. Understanding these partitions helps developers place new code appropriately and navigate existing code.

## Overview

| Partition | Purpose | Users | Data Freshness |
|-----------|---------|-------|----------------|
| **Core** | Fundamental gameplay | All users | Real-time |
| **Sandbox** | Predictive rankings and simulations | All users | Near real-time |
| **Lab** | Strategy research and model tuning | Admin only | Batch/offline |

## Core

The essential product. If nothing else existed, this is the app.

**What it does:**
- User authentication and permissions
- Investment pool (Calcutta) management
- Tournament bracket tracking
- Entry creation and team bidding
- Scoring and leaderboards
- Payout calculations

**Key question:** "Is this required for users to play the game?"

### Core Services (`internal/app/`)

| Service | Description |
|---------|-------------|
| `auth/` | Authentication and sessions |
| `bracket/` | Tournament bracket state and game outcomes |
| `calcutta/` | Investment pool rules, entries, payouts |
| `scoring/` | Point calculations and leaderboards |
| `tournament/` | Tournament and team management |
| `school/` | School/team metadata |

### Core Database Tables (`core.*`)

- `calcuttas`, `entries`, `entry_teams`, `payouts`
- `tournaments`, `teams`, `team_kenpom_stats`
- `seasons`, `schools`, `competitions`
- `calcutta_snapshots`, `calcutta_snapshot_*` (historical records)

### Core Frontend Routes

- `/` - Home
- `/calcuttas/*` - Pool management and entry views
- `/rules` - Game rules

---

## Sandbox

Predictions and simulations that enhance the core experience. Shows users how things *might* unfold based on Monte Carlo simulations using KenPom rankings.

**What it does:**
- Simulates tournament outcomes (thousands of iterations)
- Generates predictive entry rankings
- Shows probability distributions for placements
- Allows viewing "what-if" scenarios with different tournament states

**Key question:** "Does this help users understand their predicted standing?"

### Sandbox Services (`internal/app/`)

| Service | Description |
|---------|-------------|
| `simulate_tournaments/` | Monte Carlo tournament simulation |
| `calcutta_evaluations/` | Evaluate entries against simulated outcomes |
| `simulation_artifacts/` | Export simulation results to files |
| `simulation_game_outcomes/` | Game outcome probability specs |
| `tournament_simulation/` | Core simulation logic |
| `predicted_game_outcomes/` | KenPom-based game predictions |

### Sandbox Database Tables (`derived.*`)

- `simulation_runs`, `simulation_run_batches`
- `simulated_calcuttas`, `simulated_entries`, `simulated_entry_teams`
- `simulated_calcutta_payouts`, `simulated_calcutta_scoring_rules`
- `game_outcome_runs`
- `run_jobs`, `run_progress_events`, `run_artifacts`

### Sandbox Frontend Routes

- `/sandbox/*` - Simulation cohorts and results
- `/sandbox/cohorts/*` - Cohort management
- `/sandbox/simulated-calcuttas/*` - Simulated pool views
- `/sandbox/simulation-runs/*` - Individual run details
- `/runs/*` - Public predictive rankings view

### Sandbox Workers

- `simulation_worker.go` - Processes simulation runs
- `game_outcome_worker.go` - Generates predicted game outcomes
- `calcutta_evaluation_worker.go` - Evaluates calcuttas against simulations

---

## Lab

Private admin-only section for researching investment strategies. Uses historical Calcuttas to evaluate and tune models with the goal of improving our chances of winning.

**What it does:**
- Analyzes historical investment patterns
- Tests bidding strategies against past tournaments
- Evaluates market share prediction models
- Tunes optimizer parameters
- Generates candidate entry portfolios

**Key question:** "Does this help us build better investment strategies?"

### Lab Services (`internal/app/`)

| Service | Description |
|---------|-------------|
| `lab_candidates/` | Candidate entry management |
| `suite_evaluations/` | Evaluation suite execution |
| `suite_scenarios/` | Test scenario definitions |
| `synthetic_scenarios/` | Generated test calcuttas |
| `strategy_runs/` | Strategy generation tracking |
| `recommended_entry_bids/` | Bid recommendation engine |
| `ml_analytics/` | ML model integration |
| `model_catalogs/` | Model registry |
| `algorithm_registry/` | Algorithm definitions |
| `analytics/` | Historical analysis |

### Lab Database Tables (`derived.*`)

- `candidates`, `candidate_bids`
- `suites`, `suite_scenarios`, `suite_executions`, `suite_calcutta_evaluations`
- `synthetic_calcuttas`, `synthetic_calcutta_cohorts`, `synthetic_calcutta_candidates`
- `algorithms`
- `market_share_runs`
- `entry_evaluation_requests`

### Lab Frontend Routes

- `/lab/*` - Lab home and navigation
- `/lab/candidates/*` - Candidate entry analysis
- `/lab/advancements/*` - Tournament advancement models
- `/lab/investments/*` - Investment strategy models

### Lab Workers

- `strategy_generation_worker.go` - Generates recommended bids
- `market_share_worker.go` - Computes market share predictions
- `entry_evaluation_worker.go` - Evaluates entry performance

---

## Shared Infrastructure

Some code serves multiple partitions:

| Component | Used By | Location |
|-----------|---------|----------|
| `workers/` utilities | Sandbox, Lab | `internal/app/workers/` |
| `bundles/` import/export | Core, Lab | `internal/bundles/` |
| Database pool | All | `internal/db/` |
| HTTP middleware | All | `internal/transport/httpserver/middleware/` |

---

## Guidance for New Code

**When adding a new feature, ask:**

1. **Is it required for gameplay?** → Core
2. **Does it show predictions to users?** → Sandbox
3. **Does it help tune our strategy?** → Lab

**Naming conventions:**

- Sandbox tables: `simulated_*`, `simulation_*`
- Lab tables: `suite_*`, `synthetic_*`, `candidate*`, `*_runs` (for ML pipelines)

**Current state vs. ideal:**

The codebase currently uses two database schemas (`core`, `derived`) rather than three. The `derived` schema contains both Sandbox and Lab tables. Future refactoring may separate these, but the logical partition should guide code organization regardless of physical schema.

---

## Related Documentation

- `docs/design/core_vs_lab_architecture.md` - Database schema strategy
- `docs/design/lab_and_sandbox_refactor.md` - Refactoring plans
- `docs/runbooks/data_science_sandbox.md` - Data science workflow
