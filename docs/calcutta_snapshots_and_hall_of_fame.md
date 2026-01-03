# Calcutta Snapshots + Investment Pools + Hall of Fame

## Goals
- Enable evaluations of:
  - the real calcutta "as-is"
  - variant calcuttas (exclude an entry, inject a synthetic entry, etc.)
- Define a meaningful Hall of Fame scope via **investment pools**.
- Support luck and sabermetrics style metrics.

## Calcutta snapshots
A **calcutta snapshot** is an immutable representation of:
- entries + their team bids (in points)
- payout structure (in cents)
- scoring rules reference

Terminology:
- "Simulation" refers to tournament simulations.
- "Evaluation" refers to applying tournament simulations to a calcutta snapshot.

Use cases:
- Evaluate the real calcutta at different tournament times (via tournament snapshots)
- Evaluate counterfactual variants:
  - exclude a specific entry
  - swap in a user strategy (synthetic entry)
  - historical what-if analysis

Calcutta snapshots are inputs to calcutta evaluation runs.

Proposed data shape:
- `calcutta_snapshots`
  - `id`
  - `base_calcutta_id`
  - `snapshot_type` (as_is, variant)
  - `created_at`
  - `description`
- `calcutta_snapshot_entries`
  - `calcutta_snapshot_id`
  - `entry_id` (or `entry_key` if we support anonymous strategies)
  - `display_name`
- `calcutta_snapshot_entry_teams`
  - `calcutta_snapshot_id`
  - `entry_id`
  - `team_id`
  - `bid_points`

Invariant:
- Snapshots are immutable; new variants create new snapshot IDs.

## Calcutta evaluation outputs
Evaluation outputs are stored in cached derived tables:
- `entry_simulation_outcomes`: per-entry, per-simulation rows (the discrete distribution samples)
- `entry_performance`: per-entry aggregates across simulations (e.g., mean/median, p_top1, p_in_money)

## Investment pools
A pool is a long-running group of people who play together year over year.

Use cases:
- Pool-scoped hall-of-fame (meaningful, not gameable)
- "Invite everyone back" via memberships

Proposed data shape:
- `investment_pools`
  - `id`, `name`, `created_at`, `owner_id`
- `investment_pool_memberships`
  - `pool_id`, `person_id`, `role`, `active_from`, `active_to`
- `investment_pool_calcuttas`
  - `pool_id`, `calcutta_id`

## Hall of Fame metrics
### Predicted metrics (sabermetrics)
Computed from evaluation runs.
- **GM-of-the-year**: highest `expected_normalized_payout_mean` (pre-tournament snapshot)
- **Best entry of all time (expected)**: best expected normalized payout at tournament start

### Realized metrics (rings)
Computed from finalized tournament outcomes.
- **Championships / rings**: who actually won / cashed / paid out

### Luck
Luck is signed; do not take absolute value.
- `luck_normalized_payout = actual_normalized_payout - expected_normalized_payout_mean`
  - positive: overachiever
  - negative: underachiever

To enable this, we need realized outcomes stored:
- `entry_realized_outcomes` (per calcutta)
  - actual payout (cents) and/or normalized payout
  - final rank
  - scored points (if needed)

For now, realized outcomes can be computed on demand (function/view) from core facts.

## Notes
- Hall of fame UI can migrate from the admin console to pool-scoped pages.
- Keep units strict in storage/API (`*_points` vs `*_cents`), even if UI labels points as "$".
