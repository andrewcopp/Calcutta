# Verify Bundles

Verifies that bundle archives on disk match the current database state.

## Run

From the repo root:

```bash
go run ./backend/cmd/tools/verify-bundles --in ./exports/bundles
```

## Flags

- `--in` (default: `./exports/bundles`) Input bundles directory

## Output

- Exits non-zero if mismatches are found.
- Logs mismatches to stderr.

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).
