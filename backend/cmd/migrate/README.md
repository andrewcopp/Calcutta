# Migrate

Database schema migration runner.

## Run

From the repo root:

```bash
# Apply all up migrations
go run ./backend/cmd/migrate -up

# Roll back the most recent migration
go run ./backend/cmd/migrate -down

# Force schema version (clears dirty state)
go run ./backend/cmd/migrate -force <version>
```

## Configuration

Uses `backend/internal/db.Initialize()` which loads config via `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Writes to the database schema via migrations in `backend/migrations/schema/`.
