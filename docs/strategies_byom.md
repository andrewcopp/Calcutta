# Strategies (BYOM / Sandbox) and Evaluation API

## Goal
Enable open-source-friendly and premium-friendly workflows:
- Users can download historical datasets
- Run their own models/optimizers locally
- Upload a "strategy" (optimized entry)
- Core evaluates it using official simulation + evaluation harness

"Sandbox" and "Lab" are competing names for the feature.

## Key idea
Treat user uploads as **strategies** (inputs), not privileged ML artifacts.

## Export API (download datasets)
Capabilities:
- Download tournaments (teams, bracket structure, games, scoring rules)
- Download calcuttas (entries + bids, payout structure)
- Download evaluation results (optional)

Constraints:
- Authenticated via API key
- Rate limited
- Redact private pool data unless authorized

## Strategy API (upload)
A strategy is a proposed portfolio/entry.

Proposed fields:
- `strategy_id`
- `created_by`
- `created_at`
- `base_calcutta_id` or `calcutta_snapshot_id`
- `label`
- `entry_portfolio`:
  - list of `(team_id, bid_points)`

Rules:
- bids are in points (`bid_points`)
- must satisfy constraints (min/max teams, max per team, total budget)

## Evaluation of strategies
Evaluation is done by creating:
- a **calcutta snapshot** variant that injects the strategy
- an **evaluation run** referencing a chosen simulation batch

This yields:
- distribution of outcomes
- expected normalized payout
- leaderboard position vs the pool’s historical field

Compute policy:
- Airflow runs evaluations on a schedule (not per-request)
- results are cached and served by the core app

## Units (strict)
- in-game: points (`*_points`)
- real-world: cents (`*_cents`)
- normalized payout: unitless

UI may label points as "$" for investment framing, but APIs must remain explicit.

## Premium hooks (optional)
- strategy quotas
- increased simulation counts
- private strategies
- pool-scoped offline tournaments
