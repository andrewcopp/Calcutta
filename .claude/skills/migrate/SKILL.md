---
name: migrate
description: Create a new database migration file pair (up and down)
argument-hint: "[migration-name]"
allowed-tools: Bash, Write
---

# Skill: Create Migration

Create a new database migration file pair (up and down).

## Usage

```
/migrate <name>
```

Where `<name>` is a short snake_case description (e.g., `add_users_table`, `drop_legacy_columns`).

## Instructions

1. Generate a timestamp using the current time in `YYYYMMDDHHMMSS` format (include seconds).
2. Create two files in `backend/migrations/schema/`:
   - `{timestamp}_{name}.up.sql`
   - `{timestamp}_{name}.down.sql`
3. The **up** migration should contain a header comment and the SQL:

```sql
-- Migration: {name}
-- Created: {human-readable timestamp}

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- TODO: Add migration SQL here
```

4. The **down** migration should contain a matching rollback:

```sql
-- Rollback: {name}
-- Created: {human-readable timestamp}

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- TODO: Add rollback SQL here
```

5. After creating the files, print the file paths so the user can navigate to them.
6. Remind the developer to run `make sqlc-generate` after applying schema changes.

## New Table Template

When creating a new table, use this expanded template for the up migration:

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

CREATE TABLE IF NOT EXISTS {schema}.{table_name} (
    id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    -- columns here
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- updated_at trigger (every table needs this)
CREATE TRIGGER trg_{schema}_{table_name}_updated_at
    BEFORE UPDATE ON {schema}.{table_name}
    FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

-- Foreign key constraints
ALTER TABLE {schema}.{table_name}
    ADD CONSTRAINT {table_name}_{fk_column}_fkey
    FOREIGN KEY ({fk_column}) REFERENCES {other_schema}.{other_table}(id);

-- CHECK constraints for data integrity
ALTER TABLE {schema}.{table_name}
    ADD CONSTRAINT ck_{schema}_{table_name}_{column}_positive CHECK ({column} > 0);

-- Indexes: FK columns, filtered columns, soft-delete-aware partial indexes
CREATE INDEX IF NOT EXISTS idx_{table_name}_active
    ON {schema}.{table_name}(id) WHERE (deleted_at IS NULL);

CREATE INDEX IF NOT EXISTS idx_{table_name}_{fk_column}
    ON {schema}.{table_name}({fk_column}) WHERE (deleted_at IS NULL);
```

And the corresponding down migration:

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE IF EXISTS {schema}.{table_name}
    DROP CONSTRAINT IF EXISTS {table_name}_{fk_column}_fkey;
ALTER TABLE IF EXISTS {schema}.{table_name}
    DROP CONSTRAINT IF EXISTS ck_{schema}_{table_name}_{column}_positive;
DROP TRIGGER IF EXISTS trg_{schema}_{table_name}_updated_at ON {schema}.{table_name};
DROP TABLE IF EXISTS {schema}.{table_name};
```

## Alter Table Template

When adding columns to an existing table:

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Add column
ALTER TABLE {schema}.{table_name}
    ADD COLUMN IF NOT EXISTS {column} {type} {constraints};

-- Add foreign key constraint (if applicable)
ALTER TABLE {schema}.{table_name}
    ADD CONSTRAINT {table_name}_{column}_fkey
    FOREIGN KEY ({column}) REFERENCES {other_schema}.{other_table}(id);

-- Add CHECK constraint (if applicable)
ALTER TABLE {schema}.{table_name}
    ADD CONSTRAINT ck_{schema}_{table_name}_{column}_positive CHECK ({column} > 0);

-- Index the new column (if FK or frequently filtered)
CREATE INDEX IF NOT EXISTS idx_{table_name}_{column}
    ON {schema}.{table_name}({column}) WHERE (deleted_at IS NULL);
```

And the corresponding down migration:

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

DROP INDEX IF EXISTS {schema}.idx_{table_name}_{column};
ALTER TABLE IF EXISTS {schema}.{table_name}
    DROP CONSTRAINT IF EXISTS ck_{schema}_{table_name}_{column}_positive;
ALTER TABLE IF EXISTS {schema}.{table_name}
    DROP CONSTRAINT IF EXISTS {table_name}_{column}_fkey;
ALTER TABLE {schema}.{table_name}
    DROP COLUMN IF EXISTS {column};
```

## Rename / Drop Column Template

When renaming a column:

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE {schema}.{table_name}
    RENAME COLUMN {old_name} TO {new_name};
```

Down migration (reverse the rename):

```sql
SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE {schema}.{table_name}
    RENAME COLUMN {new_name} TO {old_name};
```

When dropping a column, the down migration must recreate it with the original type and constraints. Be explicit:

```sql
-- Up: drop the column
ALTER TABLE {schema}.{table_name}
    DROP COLUMN IF EXISTS {column};

-- Down: recreate with original type and constraints
ALTER TABLE {schema}.{table_name}
    ADD COLUMN IF NOT EXISTS {column} {original_type} {original_constraints};
```

## Rules

- **DO NOT wrap migrations in BEGIN/COMMIT.** The golang-migrate postgres driver wraps each migration file in a transaction automatically. Adding your own `BEGIN`/`COMMIT` creates nested transaction issues and can leave migrations in a dirty state.
- **Exception:** If a migration contains statements that cannot run inside a transaction (e.g., `CREATE INDEX CONCURRENTLY`, `ALTER TYPE ... ADD VALUE`), it must be handled specially. Split it into a separate migration and coordinate with the team.
- **Always schema-qualify** all table, index, trigger, and function names (e.g., `core.users`, not `users`). Tables belong to `core`, `derived`, `lab`, or `archive` schemas.
- **Always use `public.uuid_generate_v4()`** (fully qualified). Because `search_path = ''`, an unqualified `uuid_generate_v4()` will fail at runtime. The extension is installed in the `public` schema.
- **Use defensive SQL:** `IF EXISTS`, `IF NOT EXISTS`, `DROP CONSTRAINT IF EXISTS`.
- Never modify existing migration files -- create new ones to fix issues.
- Use snake_case for the name portion of the filename.
- The timestamp must reflect the actual current time (use `date` command), not a hardcoded value.
- Every new table must have a `set_updated_at` trigger, `created_at`/`updated_at` columns, and a `deleted_at` column for soft deletes.
- Index all foreign key columns. Use partial indexes with `WHERE (deleted_at IS NULL)` for soft-delete-aware queries.
