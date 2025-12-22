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

The CLI supports two modes:

- `-mode export`
- `-mode baseline`

The default is `export`.

### Common flags

- `-year <YYYY>`
  - Convenience flag: resolves a single Calcutta for the given tournament year.
  - If multiple calcuttas exist for that year, the CLI errors and you must use `-calcutta-id`.
- `-calcutta-id <uuid>`
  - Selects a Calcutta directly.
- `-out <path>`
  - Writes CSV output to a file. If omitted, output is written to stdout.

You must provide **exactly one** of `-year` or `-calcutta-id`.

## Export mode

Export produces one row per tournament team in the Calcutta with:

- Team identifiers and metadata (seed, region, school)
- Actual realized points
- Total community investment and normalized bid share

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
