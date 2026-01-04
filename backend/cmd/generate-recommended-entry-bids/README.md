# Generate Recommended Entry Bids

Runs the portfolio optimizer and writes recommended bids for a given calcutta.

## Run

From the repo root:

```bash
go run ./backend/cmd/generate-recommended-entry-bids --calcutta-id <uuid>
```

## Flags

- `--calcutta-id` (required) Core calcutta UUID
- `--run-key` (optional) Run key (defaults to random UUID)
- `--name` (optional) Human-readable run name
- `--optimizer` (default: `minlp_v1`) Optimizer key
- `--budget` (optional) Budget points (default: calcutta `budget_points`)
- `--min-teams` (optional) Default: calcutta `min_teams`
- `--max-teams` (optional) Default: calcutta `max_teams`
- `--min-bid` (default: `1`) Min bid points
- `--max-bid` (optional) Default: calcutta `max_bid`

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Writes a new strategy generation run and recommended entry bids to Postgres.
