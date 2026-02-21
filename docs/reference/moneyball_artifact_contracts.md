# Moneyball Artifact Contracts

This document defines the canonical contracts for Moneyball pipeline artifacts.

The purpose of a contract is to make each pipeline boundary explicit and testable:
- artifacts have stable names
- artifacts have required columns
- artifacts have invariants (probabilities sum to 1, etc.)
- runners can validate artifacts as quality gates

## General rules
- Artifacts are written as **Parquet** by default.
- Artifact names use the convention: `{observed|predicted|simulated|recommended}_{domain}_{noun}`.
- Each artifact contract specifies:
  - required columns
  - value invariants
  - upstream inputs

## Artifact: predicted_game_outcomes

### What it represents
A bracket-consistent set of matchup probabilities, including:
- unconditional matchup probability (`p_matchup`)
- conditional win probabilities (`p_team1_wins_given_matchup`, `p_team2_wins_given_matchup`)

### File
- `predicted_game_outcomes.parquet`

### Required columns
- `game_id`
- `round`
- `round_order`
- `sort_order`
- `team1_key`
- `team1_school_name`
- `team2_key`
- `team2_school_name`
- `p_matchup`
- `p_team1_wins_given_matchup`
- `p_team2_wins_given_matchup`

### Invariants
- `p_matchup` is in `[0, 1]`
- `p_team1_wins_given_matchup` is in `[0, 1]`
- `p_team2_wins_given_matchup` is in `[0, 1]`
- For every row: `p_team1_wins_given_matchup + p_team2_wins_given_matchup == 1` (within tolerance)

### Upstream inputs
- `<snapshot_dir>/games.parquet`
- `<snapshot_dir>/teams.parquet`

## Artifact: predicted_market_share

### What it represents
A normalized market-belief distribution over teams: predicted share of total auction spend.

### File
- `predicted_market_share.parquet`

### Required columns
- `predicted_market_share`

Optional identity columns (if present they are carried through)
- `snapshot`
- `tournament_key`
- `calcutta_key`
- `team_key`
- `school_name`
- `school_slug`
- `seed`
- `region`
- `kenpom_net`

### Invariants
- `predicted_market_share` is finite
- `predicted_market_share` is non-negative
- `predicted_market_share` sums to 1 across all rows (within tolerance)

### Upstream inputs
- Uses the Option-A `out_root` layout (snapshot directory names).
- Requires at least:
  - `<snapshot>/derived/team_dataset.parquet`
  - `<snapshot>/entries.parquet`
  - `<snapshot>/entry_bids.parquet`
  - `<snapshot>/teams.parquet`
  - `<snapshot>/payouts.parquet`

The runner fingerprints these input files for:
- the prediction snapshot
- the training snapshots
- the previous-year snapshot (only if present and needed for last-year features)

## Artifact: recommended_entry_bids

### What it represents
A recommended set of integer-dollar bids for a single simulated entry.

This artifact is the first “decision” artifact: it consumes market beliefs and
team value estimates and outputs a concrete portfolio.

### File
- `recommended_entry_bids.parquet`

### Required columns
- `team_key`
- `bid_amount_dollars`

Optional modeling/debug columns (if present they are carried through)
- `expected_team_points`
- `predicted_team_total_bids`
- `predicted_market_share`
- `score`

### Invariants
- `bid_amount_dollars` is an integer
- `bid_amount_dollars` is non-negative
- `team_key` is non-empty
- `team_key` is unique (no duplicate team rows)

The following invariants are stage-config dependent (and are validated by the
runner/optimizer rather than the static artifact contract):
- bids sum to `budget_dollars`
- number of teams is within `[min_teams, max_teams]`
- each bid is within `[min_bid_dollars, max_per_team_dollars]`

### Upstream inputs
- `predicted_game_outcomes.parquet`
- `predicted_market_share.parquet`

## Artifact: simulated_entry_outcomes

### What it represents
Monte Carlo simulated distribution of outcomes for the simulated entry defined
by `recommended_entry_bids`, against the observed market bids.

### Files
- `simulated_entry_outcomes.parquet` (summary)
- `simulated_entry_outcomes_sims.parquet` (optional, per-simulation rows)

### Required columns (summary)
- `entry_key`
- `sims`
- `seed`
- `budget_dollars`
- `mean_payout_cents`
- `mean_total_points`
- `mean_finish_position`
- `p_top1`
- `p_top3`
- `p_top6`
- `p_top10`

### Invariants
- `sims` is positive
- `budget_dollars` is positive
- probabilities `p_top*` are in `[0, 1]`
- payout fields are non-negative

### Upstream inputs
- Snapshot files:
  - `games.parquet`
  - `teams.parquet`
  - `payouts.parquet`
  - `entry_bids.parquet`
- Artifacts:
  - `predicted_game_outcomes.parquet`
  - `recommended_entry_bids.parquet`

## Artifact: investment_report

### What it represents
Consolidated analysis report aggregating portfolio allocation and simulated
performance into actionable investment insights.

### File
- `investment_report.parquet`

### Required columns
- `snapshot_name`
- `budget_dollars`
- `n_sims`
- `seed`
- `portfolio_team_count`
- `portfolio_total_bids`
- `mean_expected_payout_cents`
- `mean_expected_points`
- `mean_expected_finish_position`
- `p_top1`
- `p_top3`
- `p_top6`
- `p_top10`
- `expected_roi`
- `portfolio_concentration_hhi`
- `portfolio_teams_json`

### Invariants
- `portfolio_team_count` is positive
- `portfolio_total_bids` equals `budget_dollars`
- `expected_roi` is non-negative
- `portfolio_concentration_hhi` is in `[0, 1]`
- probabilities `p_top*` are in `[0, 1]`

### Upstream inputs
- Artifacts:
  - `predicted_game_outcomes.parquet`
  - `predicted_market_share.parquet`
  - `recommended_entry_bids.parquet`
  - `simulated_entry_outcomes.parquet`
