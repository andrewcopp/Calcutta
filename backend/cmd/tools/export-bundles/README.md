# Export Bundles

Exports bundle archives from the database to a local directory.

## Run

From the repo root:

```bash
go run ./backend/cmd/tools/export-bundles

# Custom output directory
go run ./backend/cmd/tools/export-bundles --out /tmp/calcutta-bundles
```

## Flags

- `--out` (default: `./exports/bundles`) Output directory

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Writes files to the output directory.
