# Investment Prediction Experiments

## Goal
Predict *market investment* (team bid share) well enough to identify *overbid / underbid* teams, especially among high-quality teams.

## Current evaluation runner
- `backend/scripts/run_invest_eval.sh`
- Output: CSV containing per-year metrics and overall aggregated rows.

## Metrics to track
### Primary (north star)
- `bid_share_resid_mae_std`

### Secondary
- `bid_share_resid_mae`
- `bid_share_mae_equal_seed`
- `bid_share_mae_seed_norm`
- `bid_share_mae`

### Product-facing
- `underbid_top3_hit_rate_top16`
- `underbid_top3_hit_rate_seeds_1_to_8`
- `overbid_top3_hit_rate_top16`
- `overbid_top3_hit_rate_seeds_1_to_8`

## Baseline models (existing)
- `seed`
- `seed-pod`
- `seed-kenpom-delta`
- `seed-kenpom-rank`
- `kenpom-rank`
- `kenpom-score`

## Model backlog
- `ridge` (Go-native)
- `elastic-net` (Go-native)
- `monotone-calibration` (isotonic) on:
  - KenPom percentile
  - predicted returns (expected points share)

## Feature backlog
### Strength / efficiency
- KenPom percentile (overall and within-seed)
- KenPom z-score within seed line
- KenPom AdjO / AdjD / SOS / Luck (if present in DB)

### Returns-derived
- `expected_points_share` (predicted returns)
- `P(reach Sweet16)` / `P(reach Elite8)` / `P(win 1 game)` using bracket win probabilities
- sigma sensitivity features (smaller sigma variants)

### Market behavior / narratives
- prior-year seed
- prior-year tournament wins / points
- rolling tournament appearances (last N years)
- rolling tournament wins (last N years)
- rolling "value produced" vs price (prior-year residuals)

## Experiment template
- **Hypothesis:**
- **Feature set:**
- **Model:**
- **Train window:**
- **Result:** (north star + product-facing hit rates)
- **Notes / next steps:**
