# Data Science

Small Python utilities used to ingest/export snapshots and generate canonical datasets.

## Setup

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Scripts

- `ingest_snapshot.py`
  - Ingest a snapshot export from the backend admin tools.
- `derive_canonical.py`
  - Build a canonical dataset from ingested files.
- `evaluate_harness.py`
  - Run leakage-safe baseline evaluation across snapshots.
- `backtest_scaffold.py`
  - Build a feasible portfolio under auction constraints and compute realized + expected payout/ROI.
- `calibrate_kenpom_scale.py`
  - Fit a global KenPom win probability scale from historical outcomes.
- `investment_report.py`
  - Generate per-snapshot investment reports (portfolio + realized/expected evaluation) with staged artifacts and caching.
  - See `calcutta_ds/investment_report/README.md` for details on outputs and CLI flags.

## How to run

### 1) Export + ingest a snapshot (Option A layout)

- Generate an API key in the admin console.
- Set it as:

```bash
export CALCUTTA_API_KEY="..."
```

Then ingest:

```bash
python ingest_snapshot.py \
  --base-url http://localhost:8080 \
  --tournament-id <TOURNAMENT_ID> \
  --calcutta-id <CALCUTTA_ID> \
  --api-key "$CALCUTTA_API_KEY" \
  --out-root ./out \
  --update-latest
```

The snapshot will be written to `./out/<snapshot>/` and `./out/LATEST` will contain the snapshot directory name.

### 2) Derive canonical tables

```bash
SNAP=./out/$(cat ./out/LATEST)
python derive_canonical.py "$SNAP"
```

This writes `derived/team_features.parquet`, `derived/team_market.parquet`, and `derived/team_dataset.parquet`.

### 3) Evaluate baselines across snapshots

```bash
python evaluate_harness.py ./out --out ./out/eval_report.json
```

### 4) Backtest scaffold (realized payout/ROI)

```bash
SNAP=./out/$(cat ./out/LATEST)
python backtest_scaffold.py "$SNAP" --out "$SNAP/derived/backtest.json"
```

### 5) Expected payout/ROI (Monte Carlo)

```bash
SNAP=./out/$(cat ./out/LATEST)
python backtest_scaffold.py "$SNAP" \
  --expected-sims 1000 \
  --expected-seed 1 \
  --out "$SNAP/derived/backtest_with_expected.json"
```

### 6) Calibrate KenPom scale (recommended) + use it

```bash
python calibrate_kenpom_scale.py ./out --out ./out/kenpom_scale.json
```

Then:

```bash
SNAP=./out/$(cat ./out/LATEST)
python backtest_scaffold.py "$SNAP" \
  --expected-sims 1000 \
  --expected-seed 1 \
  --kenpom-scale-file ./out/kenpom_scale.json \
  --out "$SNAP/derived/backtest_with_expected.json"
```

## Notes

- This folder intentionally stays lightweight.
- For more context, see `docs/runbooks/data_science_sandbox.md`.
