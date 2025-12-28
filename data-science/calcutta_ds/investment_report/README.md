# Investment Report Harness

This harness generates per-snapshot investment strategy outputs and backtest evaluation, with optional staged artifacts and caching to make iteration fast and human-readable.

Entry point:

```bash
PYTHONPATH=data-science python data-science/investment_report.py <out_root> [flags]
```

Where `<out_root>` is an Option-A snapshot root containing snapshot directories (e.g. `data-science/out/2019/`, `data-science/out/2025/`).

## What it does (high level)

For each snapshot year:

1) **Predict market investment**
- Predicts `predicted_team_share_of_pool` per team using prior snapshots (`--train-mode past_only`).
- Converts share -> predicted dollars via `predicted_team_total_bids = predicted_team_share_of_pool * n_entries * budget`.

2) **Predict team scoring (expected points)**
- Uses Monte Carlo tournament simulation to estimate `expected_team_points` per team.

3) **Compute ROI diagnostics**
- Builds ROI-style tables like `predicted_points_per_dollar_at_min_bid` using the predicted market totals.

4) **Build a portfolio**
- Chooses bids subject to constraints:
  - `budget` (default 100)
  - `min_teams` / `max_teams` (default 3..10)
  - `min_bid` (default 1)
  - `max_per_team` (default 50)

Supported allocation modes:
- `equal` (baseline)
- `expected_points` (allocate bids by expected points)
- `greedy` (greedy optimization)
- `knapsack` (DP / classic knapsack global optimum for integer-dollar bids)

5) **Evaluate vs the market**
- **Realized**: performance against the actual market and actual tournament results for that year.
- **Expected**: simulated performance distribution vs the market.

## Output report JSON

The main output is a single JSON report (default: `<out_root>/report.json` or `--out <path>`):

- `years[]`
  - `portfolio`: chosen team bids
  - `strategy`: constraint settings and allocation mode
  - `market_model`: training snapshots + ridge alpha
  - `realized`: realized points/finish/payout
  - `expected` (if `--expected-sims > 0`): mean/percentiles + `p_top1/p_top3/p_top6/p_top10`
  - `debug` (if `--debug-output`): ranked tables and diagnostics

## Staged artifacts (recommended for debugging)

Use `--artifacts-dir <dir>` to write small per-snapshot CSV/JSON artifacts under:

`<artifacts_dir>/<snapshot>/`

### Stage files

- `scores.csv`
  - Team expected scoring fields.
- `predicted_investments.csv`
  - Team predicted market share and predicted total bids.
- `predicted_roi.csv`
  - ROI table sorted by `predicted_points_per_dollar_at_min_bid`.
- `portfolio.json`
  - Final selected team bids (human-readable).
- `expected.json`
  - Expected performance distribution summary (Monte Carlo).
- `realized.json`
  - Realized finish position, payout, and ROI.
- `standings.csv`
  - Full standings including the simulated entry.

### Cache files (speed up reruns)

These are written automatically when `--artifacts-dir` is set:

- `predicted_market.json`
- `predicted_market_meta.json`
- `team_mean_points.json`
- `team_mean_points_meta.json`

If you re-run with `--use-cache`, the harness will reuse these cached artifacts when the corresponding `*_meta.json` matches the current run’s config.

## Snapshot filtering (run only the years you want)

To speed iteration, you can restrict which snapshots are run:

- Single year:

```bash
--only-snapshot 2025
```

- Multiple years:

```bash
--snapshots 2023,2024,2025
```

You cannot set both flags simultaneously.

## Common commands

### 1) Fast single-year debug run (recommended)

```bash
PYTHONPATH=data-science python data-science/investment_report.py data-science/out \
  --train-mode past_only \
  --allocation-mode knapsack \
  --only-snapshot 2025 \
  --expected-sims 200 \
  --expected-seed 1 \
  --debug-output \
  --artifacts-dir data-science/out/artifacts \
  --use-cache \
  --out data-science/out/report_2025_debug.json
```

### 2) Full multi-year run (no filtering)

```bash
PYTHONPATH=data-science python data-science/investment_report.py data-science/out \
  --train-mode past_only \
  --allocation-mode knapsack \
  --expected-sims 2000 \
  --expected-seed 1 \
  --debug-output \
  --artifacts-dir data-science/out/artifacts \
  --use-cache \
  --out data-science/out/report.json
```

## Notes

- Excluding your own entry from market data is supported via repeated `--exclude-entry-name` flags (default excludes `Andrew Copp`).
- `--expected-use-historical-winners` is leaky for true pre-tournament expectations; use only for sanity checks.
