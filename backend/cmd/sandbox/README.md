# Calcutta Data Science Sandbox (CLI)

This package contains a small Go CLI intended for repeatable data exports and baseline modeling against the Calcutta Postgres database.

It is designed to:

- Export a per-team dataset (points + community investment) for a single Calcutta.
- Run simple baseline models (currently: seed-only) to establish a measurement baseline before adding richer features.

## Philosophy

This sandbox is meant to support research that focuses on **expected value** and **uncertainty**, not “overfitting to history” (e.g., cinderella runs / bad beats).

Baseline models are deliberately simple and should be treated as yardsticks.

## Prerequisites

- Go installed (project uses Go `1.24`)
- Postgres running and seeded with Calcutta data
- A valid `DATABASE_URL` (or component env vars used by `run_sandbox.sh`)

## How to run

The recommended entrypoint is the wrapper script:

```bash
./backend/scripts/run_sandbox.sh [flags]
```

For the common “end-to-end” workflow (all years, exclude "Andrew Copp", output report), use:

```bash
./backend/scripts/run_sandbox_report.sh
```

The wrapper script:

- Loads env vars from `backend/.env` and `./.env` if present
- If `DATABASE_URL` is not set, it builds one from:
  - `DB_HOST` (defaults to `localhost` if unset or `db`)
  - `DB_PORT` (default `5432`)
  - `DB_NAME` (default `calcutta`)
  - `DB_USER` (default `calcutta`)
  - `DB_PASSWORD` (default `calcutta`)

### Note about `/tmp` on macOS

On macOS, `/tmp` is typically a symlink to `/private/tmp`.

So an output like `/tmp/teams_2025.csv` will actually live at `/private/tmp/teams_2025.csv`.

## Modes

The CLI supports these modes:

- `-mode export`
- `-mode baseline`
- `-mode simulate`
- `-mode backtest`
- `-mode report`

The default is `export`.

### Common flags

- `-year <YYYY>`
  - Convenience flag: resolves a single Calcutta for the given tournament year.
  - If multiple calcuttas exist for that year, the CLI errors and you must use `-calcutta-id`.
- `-calcutta-id <uuid>`
  - Selects a Calcutta directly.
- `-out <path>`
  - Writes CSV output to a file. If omitted, output is written to stdout.
- `-exclude-entry-name <name>`
  - Excludes bids from entries with this exact name when computing investment totals/shares.
  - Example: `-exclude-entry-name "Andrew Copp"`
  - This is useful for measuring cannibalization and reducing strategy leakage when training investment models.

You must provide **exactly one** of `-year` or `-calcutta-id`.

## Export mode

Export produces one row per tournament team in the Calcutta with:

- Team identifiers and metadata (seed, region, school)
- Actual realized points
- Total community investment and normalized bid share
- Total community investment and normalized bid share excluding `-exclude-entry-name` (if provided)

Example:

```bash
./backend/scripts/run_sandbox.sh \
  -mode export \
  -year 2025 \
  -out /tmp/teams_2025.csv
```

## Baseline mode

Baseline mode trains a seed-only model on historical data and evaluates it against a target Calcutta.

It writes one row per team containing:

- Actual points and predicted points (by seed)
- Actual bid share and predicted bid share (by seed)
- Actual bid share excluding `-exclude-entry-name` (if provided)
- Normalized ROI (actual vs predicted)

### Training window (`-train-years`)

Baseline supports an optional rolling training window:

- `-train-years 0` (default): use **all** historical data excluding the target calcutta
- `-train-years 1`: use only the **previous 1** tournament year
- `-train-years 2`: use only the **previous 2** tournament years
- `-train-years 3`: use only the **previous 3** tournament years

Example (all history):

```bash
./backend/scripts/run_sandbox.sh \
  -mode baseline \
  -year 2025 \
  -out /tmp/baseline_2025_all_history.csv
```

Example (rolling window):

```bash
./backend/scripts/run_sandbox.sh \
  -mode baseline \
  -year 2025 \
  -train-years 1 \
  -out /tmp/baseline_2025_train1.csv

./backend/scripts/run_sandbox.sh \
  -mode baseline \
  -year 2025 \
  -train-years 2 \
  -out /tmp/baseline_2025_train2.csv

./backend/scripts/run_sandbox.sh \
  -mode baseline \
  -year 2025 \
  -train-years 3 \
  -out /tmp/baseline_2025_train3.csv
```

Baseline prints a summary to stderr:

- `points_mae`
- `bid_share_mae`

## Simulate mode

Simulate mode generates a single “simulated entry” (a bid allocation) for a target Calcutta.

It uses:

- A returns model: seed-based expected points (trained from historical years via `-train-years`)
- A market model: baseline market investment `B_i` from the target Calcutta (optionally excluding `-exclude-entry-name`)

And then solves for an allocation that maximizes expected entry points under constraints.

### Constraints (defaults)

- Budget is `100` (whole dollars)
- Minimum teams bid on: `3`
- Maximum teams bid on: `10`
- Minimum bid per team: `1`
- Maximum bid per team: `50`

All of these are configurable with flags:

- `-budget`
- `-min-teams`, `-max-teams`
- `-min-bid`, `-max-bid`

### Why this is not greedy

Under a `max-teams` cap, a naive “take the next best marginal ROI” greedy algorithm can get stuck: you hit 10 teams, then you’re forced to keep adding dollars to the same 10 even when their marginal ROI collapses.

This sandbox uses an integer dynamic programming optimizer (knapsack-style) to avoid that lock-in and choose the best 3–10 team set and bid sizes jointly.

### Output

The CSV includes **only teams with non-zero bids**, matching how a real entry looks.

Example:

```bash
./backend/scripts/run_sandbox.sh \
  -mode simulate \
  -year 2025 \
  -exclude-entry-name "Andrew Copp" \
  -train-years 0 \
  -out /tmp/sim_entry_2025.csv
```

## Backtest mode

Backtest runs `simulate` across a year range and scores the simulated entry against realized tournament outcomes.

It outputs one row per year containing:

- Expected normalized ROI (based on the returns model)
- Realized normalized ROI (based on actual tournament points)

Years with no Calcutta are skipped.

If `-start-year` and/or `-end-year` are omitted, the sandbox will auto-detect the available year range from the database.

Example:

```bash
./backend/scripts/run_sandbox.sh \
  -mode backtest \
  -start-year 2018 \
  -end-year 2025 \
  -exclude-entry-name "Andrew Copp" \
  -train-years 0 \
  -out /tmp/backtest_2018_2025.csv
```

## Report mode

Report mode runs the end-to-end harness year-by-year and writes a human-readable report:

- The simulated entry (team + bid points)
- The simulated finish position vs the field for that year

Example:

```bash
./backend/scripts/run_sandbox.sh \
  -mode report \
  -start-year 2018 \
  -end-year 2025 \
  -exclude-entry-name "Andrew Copp" \
  -train-years 0 \
  -out /tmp/report_2018_2025.md
```

If `-start-year` and/or `-end-year` are omitted, the sandbox will auto-detect the available year range from the database.

Happy path:

```bash
./backend/scripts/run_sandbox_report.sh
```

## Troubleshooting

### "DATABASE_URL environment variable is not set"

Either:

- Set `DATABASE_URL` explicitly, or
- Provide `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, etc. and use `./backend/scripts/run_sandbox.sh`.

### "found N calcuttas for year YYYY"

Re-run with `-calcutta-id <uuid>` using one of the printed candidates.

### No output file created

- Ensure you passed `-out`.
- Ensure the parent directory exists (the CLI does not create parent directories).
- On macOS, check `/private/tmp` if you used `/tmp`.
