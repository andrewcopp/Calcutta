# Benchmark Tournament Simulations

End-to-end benchmark for tournament simulation + DB write throughput.

## Run

From the repo root:

```bash
go run ./backend/cmd/benchmark-tournament-simulations --season 2025
```

## Flags

- `--season` (required) Tournament season/year (e.g. `2025`)
- `--n-sims` (default: `100000`) Number of simulations
- `--seed` (default: `42`) Base RNG seed
- `--workers` (default: `GOMAXPROCS`) Simulation workers
- `--batch-size` (default: `1000`) Simulations per DB COPY batch
- `--probability-source-key` (default: `go_benchmark`) Stored in `derived.simulated_tournaments.probability_source_key`

## Configuration

Uses `backend/internal/db.Initialize()` which loads config via `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Creates a snapshot/batch and writes simulated team results to Postgres.
