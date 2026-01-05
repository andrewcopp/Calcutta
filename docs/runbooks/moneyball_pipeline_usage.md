# Moneyball Pipeline Usage Guide

This guide explains how to run the Moneyball pipeline stages, understand artifact outputs, and leverage manifest-based caching.

## Overview

The Moneyball pipeline consists of four stages that run sequentially:

1. **`predicted_game_outcomes`** - Monte Carlo simulation of tournament bracket outcomes
2. **`predicted_auction_share_of_pool`** - Ridge regression model predicting market bid distribution
3. **`recommended_entry_bids`** - Greedy portfolio optimizer for bid allocation
4. **`simulated_entry_outcomes`** - Monte Carlo simulation of entry performance vs market

Each stage:
- Reads from a snapshot directory (parquet tables: `games`, `teams`, `payouts`, `entry_bids`)
- Writes artifacts to a run directory under `artifacts_root`
- Uses manifest-based caching to skip redundant computation
- Validates outputs against artifact contracts

## Quick Start

### Run all stages for a snapshot

```bash
cd data-science

# Run full pipeline for 2025 snapshot
PYTHONPATH=. python -m moneyball.cli simulated-entry-outcomes \
  out/2025 \
  --snapshot-name 2025 \
  --artifacts-root out/artifacts \
  --run-id main \
  --n-sims 5000 \
  --seed 123 \
  --budget-dollars 100
```

This will:
- Run all four stages in sequence (upstream stages included by default)
- Write artifacts to `out/artifacts/main/<timestamp>/`
- Cache results based on config + input fingerprints
- Output a JSON summary with run metadata

### Run a single stage

```bash
# Run only predicted game outcomes
PYTHONPATH=. python -m moneyball.cli predicted-game-outcomes \
  out/2025 \
  --snapshot-name 2025 \
  --artifacts-root out/artifacts \
  --run-id main \
  --kenpom-scale 10.0 \
  --n-sims 5000 \
  --seed 123
```

## Pipeline Stages

### 1. predicted-game-outcomes

Predicts tournament bracket outcomes using Monte Carlo simulation over KenPom ratings.

**Inputs:**
- `<snapshot_dir>/games.parquet`
- `<snapshot_dir>/teams.parquet`

**Outputs:**
- `predicted_game_outcomes.parquet`

**Parameters:**
- `--calcutta-key` - Filter to specific Calcutta (optional)
- `--kenpom-scale` - KenPom rating scale factor (default: 10.0)
- `--n-sims` - Number of Monte Carlo simulations (default: 5000)
- `--seed` - Random seed for reproducibility (default: 123)
- `--no-cache` - Disable manifest-based caching

**Example:**
```bash
PYTHONPATH=. python -m moneyball.cli predicted-game-outcomes \
  out/2025 \
  --kenpom-scale 10.0 \
  --n-sims 10000 \
  --seed 42
```

### 2. predicted-auction-share-of-pool

Predicts market bid distribution using ridge regression trained on historical snapshots.

**Inputs:**
- Snapshot directories (current + training years)
- Each snapshot requires:
  - `derived/team_dataset.parquet`
  - `entries.parquet`
  - `entry_bids.parquet`
  - `teams.parquet`
  - `payouts.parquet`

**Outputs:**
- `predicted_auction_share_of_pool.parquet`

**Parameters:**
- `--train-snapshot` - Training snapshot names (repeatable, e.g., `--train-snapshot 2023 --train-snapshot 2024`)
- `--ridge-alpha` - Ridge regression regularization strength (default: 1.0)
- `--feature-set` - Feature set name (default: `expanded_last_year_expected`)
- `--exclude-entry-name` - Exclude entries from training (repeatable)
- `--no-cache` - Disable manifest-based caching

**Example:**
```bash
PYTHONPATH=. python -m moneyball.cli predicted-auction-share-of-pool \
  out/2025 \
  --train-snapshot 2023 \
  --train-snapshot 2024 \
  --ridge-alpha 1.0 \
  --exclude-entry-name "Andrew Copp"
```

### 3. recommended-entry-bids

Generates optimal bid allocation using greedy portfolio optimization.

**Inputs:**
- `predicted_game_outcomes.parquet` (from stage 1)
- `predicted_auction_share_of_pool.parquet` (from stage 2)

**Outputs:**
- `recommended_entry_bids.parquet`

**Parameters:**
- `--budget-dollars` - Total budget (default: 100)
- `--min-teams` - Minimum teams to bid on (default: 3)
- `--max-teams` - Maximum teams to bid on (default: 10)
- `--max-per-team-dollars` - Maximum bid per team (default: 50)
- `--min-bid-dollars` - Minimum bid amount (default: 1)
- `--predicted-total-pool-bids-dollars` - Override predicted pool size
- `--no-include-upstream` - Don't run upstream stages (default: run upstream)
- `--no-cache` - Disable manifest-based caching

**Example:**
```bash
PYTHONPATH=. python -m moneyball.cli recommended-entry-bids \
  out/2025 \
  --budget-dollars 100 \
  --min-teams 5 \
  --max-teams 8 \
  --max-per-team-dollars 30
```

### 4. simulated-entry-outcomes

Simulates entry performance distribution against observed market bids.

**Inputs:**
- `<snapshot_dir>/games.parquet`
- `<snapshot_dir>/teams.parquet`
- `<snapshot_dir>/payouts.parquet`
- `<snapshot_dir>/entry_bids.parquet`
- `predicted_game_outcomes.parquet` (from stage 1)
- `recommended_entry_bids.parquet` (from stage 3)

**Outputs:**
- `simulated_entry_outcomes.parquet` (summary statistics)
- `simulated_entry_outcomes_sims.parquet` (per-simulation details, if `--keep-sims`)

**Parameters:**
- `--calcutta-key` - Filter to specific Calcutta (optional)
- `--n-sims` - Number of Monte Carlo simulations (default: 5000)
- `--seed` - Random seed for reproducibility (default: 123)
- `--budget-dollars` - Entry budget for ownership calculation (default: 100)
- `--keep-sims` - Write per-simulation outcomes (default: false)
- `--no-include-upstream` - Don't run upstream stages (default: run upstream)
- `--no-cache` - Disable manifest-based caching

**Example:**
```bash
PYTHONPATH=. python -m moneyball.cli simulated-entry-outcomes \
  out/2025 \
  --n-sims 10000 \
  --seed 42 \
  --budget-dollars 100 \
  --keep-sims
```

## Manifest-Based Caching

The pipeline uses content-based caching to avoid redundant computation:

### How it works

1. Each stage computes a **manifest** containing:
   - Input file fingerprints (SHA256 of content)
   - Stage configuration (parameters)
   - Timestamp and metadata

2. Before running a stage, the runner checks if a previous run exists with:
   - Same input fingerprints
   - Same configuration

3. If a match is found, the stage is **skipped** and existing artifacts are reused.

### Cache keys

Each stage has different cache dependencies:

- **predicted_game_outcomes**: `games.parquet`, `teams.parquet`, `kenpom_scale`, `n_sims`, `seed`
- **predicted_auction_share_of_pool**: training snapshots, `ridge_alpha`, `feature_set`, `exclude_entry_names`
- **recommended_entry_bids**: upstream artifacts, budget constraints
- **simulated_entry_outcomes**: upstream artifacts, snapshot files, `n_sims`, `seed`, `budget_dollars`

### Controlling cache behavior

- **Use cache** (default): `--use-cache` or omit flag
- **Bypass cache**: `--no-cache` forces recomputation
- **Inspect manifests**: Check `<run_dir>/manifest.json` for cache metadata

### Example: Cache hit

```bash
# First run - computes all stages
PYTHONPATH=. python -m moneyball.cli simulated-entry-outcomes out/2025 --run-id test

# Second run - hits cache for all stages (same config)
PYTHONPATH=. python -m moneyball.cli simulated-entry-outcomes out/2025 --run-id test
```

### Example: Cache miss

```bash
# First run
PYTHONPATH=. python -m moneyball.cli predicted-game-outcomes out/2025 --seed 123

# Second run - cache miss (different seed)
PYTHONPATH=. python -m moneyball.cli predicted-game-outcomes out/2025 --seed 456
```

## Artifact Locations

Artifacts are written to:

```
<artifacts_root>/<run_id>/<timestamp>/
```

For example:

```
out/artifacts/main/2025-01-15T10-30-00Z/
├── manifest.json
├── predicted_game_outcomes.parquet
├── predicted_auction_share_of_pool.parquet
├── recommended_entry_bids.parquet
└── simulated_entry_outcomes.parquet
```

### Manifest structure

```json
{
  "run_id": "main",
  "snapshot_name": "2025",
  "timestamp": "2025-01-15T10:30:00Z",
  "stages": ["predicted_game_outcomes", "..."],
  "config": {
    "kenpom_scale": 10.0,
    "n_sims": 5000,
    "seed": 123
  },
  "inputs": {
    "games.parquet": "sha256:abc123...",
    "teams.parquet": "sha256:def456..."
  }
}
```

## Artifact Contracts

All artifacts are validated against contracts defined in `docs/reference/moneyball_artifact_contracts.md`.

Contracts specify:
- Required columns
- Value invariants (e.g., probabilities sum to 1)
- Upstream dependencies

See `moneyball/pipeline/contracts.py` for validation logic.

## Common Workflows

### Backtest across multiple years

```bash
for year in 2023 2024 2025; do
  PYTHONPATH=. python -m moneyball.cli simulated-entry-outcomes \
    out/$year \
    --snapshot-name $year \
    --run-id backtest \
    --train-snapshot 2016 \
    --train-snapshot 2017 \
    # ... (add all prior years)
done
```

### Debug a single stage

```bash
# Run only the stage you're debugging
PYTHONPATH=. python -m moneyball.cli predicted-auction-share-of-pool \
  out/2025 \
  --no-cache \
  --ridge-alpha 0.5
```

### Sensitivity analysis

```bash
# Test different KenPom scales
for scale in 5.0 10.0 15.0; do
  PYTHONPATH=. python -m moneyball.cli predicted-game-outcomes \
    out/2025 \
    --run-id "kenpom_scale_$scale" \
    --kenpom-scale $scale
done
```

## Testing

Run the full test suite:

```bash
cd data-science
python -m unittest discover -s tests -p 'test_*.py' -v
```

Run end-to-end integration tests:

```bash
python -m unittest tests.test_moneyball_pipeline_e2e -v
```

## Troubleshooting

### "Snapshot directory not found"

Ensure the snapshot directory exists and contains required parquet files:
- `games.parquet`
- `teams.parquet`
- `payouts.parquet`
- `entry_bids.parquet`

### "Artifact contract validation failed"

Check the error message for which invariant was violated. Common issues:
- Probabilities don't sum to 1 (numerical precision)
- Negative values in non-negative fields
- Missing required columns

### "No matching cached run found"

This is expected when:
- Input files have changed
- Configuration parameters have changed
- This is the first run with this config

Use `--no-cache` to force recomputation.

## Next Steps

- See `docs/reference/moneyball_artifact_contracts.md` for artifact specifications
- See `data-science/docs/design/plans.md` for future enhancements
- See `data-science/docs/design/ideas.md` for research directions
