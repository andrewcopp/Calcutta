# Calculate Simulated Calcuttas

Calculates simulated calcutta results for a tournament + strategy generation run.

## Run

From the repo root:

```bash
# Usage:
#   calculate-simulated-calcuttas <tournament_id> [run_id] [excluded_entry_name]

go run ./backend/cmd/calculate-simulated-calcuttas <tournament_id>
```

## Arguments

- `tournament_id` (required) Derived tournament id.
- `run_id` (optional) Strategy generation run key. If omitted, the latest `derived.strategy_generation_runs.run_key` is selected.
- `excluded_entry_name` (optional) Entry name to exclude. Can also be provided via `EXCLUDED_ENTRY_NAME`.

## Configuration

This binary currently builds its own DB URL from DB parts:

- `DB_USER`
- `DB_PASSWORD`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`

## Notes

- The `excluded_entry_name` value is currently only logged by this binary; it is not applied to the calculation.
