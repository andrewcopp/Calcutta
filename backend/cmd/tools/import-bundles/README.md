# Import Bundles

Imports bundle archives from a local directory into the database.

## Run

From the repo root:

```bash
# Validate bundles without committing DB writes (default)
go run ./backend/cmd/tools/import-bundles --in ./exports/bundles

# Actually import and commit
go run ./backend/cmd/tools/import-bundles --in ./exports/bundles --dry-run=false
```

## Flags

- `--in` (default: `./exports/bundles`) Input bundles directory
- `--dry-run` (default: `true`) Read and validate bundles; rollback DB writes

## Output

Prints a JSON report to stdout.

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

- With `--dry-run=true`: no committed DB changes.
- With `--dry-run=false`: writes to the database.
