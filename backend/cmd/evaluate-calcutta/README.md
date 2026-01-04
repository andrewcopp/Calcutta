# Evaluate Calcutta

Runs calcutta evaluation using an existing simulation batch.

## Run

From the repo root:

```bash
go run ./backend/cmd/evaluate-calcutta --calcutta-id <uuid>
```

## Flags

- `--calcutta-id` (required) Core calcutta UUID
- `--tournament-simulation-batch-id` (optional) Overrides the tournament simulation batch id (`derived.simulated_tournaments.id`)
- `--excluded-entry-name` (optional) Entry name to exclude (fallback: `EXCLUDED_ENTRY_NAME` env var)
- `--run-id` (optional) Run id tag for legacy compatibility (default: `go_eval`)

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Writes evaluation outputs to Postgres.
