# Calculate Simulated Calcuttas

Calculates simulated calcutta results for a tournament + strategy generation run.

## Run

From the repo root:

```bash
# Usage:
#   calculate-simulated-calcuttas --calcutta-id <calcutta_id> [--run-id <run_id>] [--excluded-entry-name <name>]

go run ./backend/cmd/calculate-simulated-calcuttas --calcutta-id <calcutta_id>
```

## Arguments

- `--calcutta-id` (required) Calcutta id.
- `--run-id` (optional) Strategy generation run key. If omitted, the latest `derived.strategy_generation_runs.run_key` is selected.
- `--excluded-entry-name` (optional) Entry name to exclude. Can also be provided via `EXCLUDED_ENTRY_NAME`.

## Configuration

This binary uses the shared config loader in `platform.LoadConfigFromEnv()`:

- Prefer `DATABASE_URL`
- Or set DB parts (`DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, optional `DB_SSLMODE`)

## Notes

- The `excluded_entry_name` value is applied when building the simulated calcutta snapshot.
