# Database Migrations

This directory contains database migrations for the Calcutta application.

## Migration Files

Migration files follow a timestamp-based naming convention:

```
YYYYMMDDHHMMSS_description.up.sql
YYYYMMDDHHMMSS_description.down.sql
```

For example:
- `20250405083239_initial_schema.up.sql`
- `20250405083239_initial_schema.down.sql`

The timestamp ensures that migrations are applied in the correct order, even when working with multiple branches or collaborators.

**Important:** Migration versions must be **unique**. Two different migration files may not share the same timestamp/version. If you create multiple migrations back-to-back, ensure each has a distinct version (e.g. wait 1 second between timestamps or manually increment by 1).

## Running Migrations

To run migrations, use the backend migration runner:

```bash
# Preferred (from repo root)
make ops-migrate

# Or run the binary directly (from repo root)
go run ./backend/cmd/migrate -up

# Roll back the last migration
go run ./backend/cmd/migrate -down
```

## Environment Variables

The migration runner loads config via `backend/internal/platform.LoadConfigFromEnv()`:

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

## Creating New Migrations

When creating new migrations:

1. Generate a new timestamp: `date +%Y%m%d%H%M%S`
2. Create two files with the same timestamp but different suffixes:
   - `{timestamp}_description.up.sql`
   - `{timestamp}_description.down.sql`
3. Write your SQL in the up file to apply the migration
4. Write your SQL in the down file to roll back the migration

If you need to create multiple migrations quickly, you can generate a base timestamp and then increment it manually for subsequent migrations (e.g. `...70000`, `...70001`, `...70002`).

## Best Practices

- Keep migrations small and focused
- Make sure down migrations properly clean up what up migrations create
- Test both up and down migrations before committing
- Never modify existing migration files after they've been applied to a database 