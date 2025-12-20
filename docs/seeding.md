# Database Seeding

This document explains the two approaches available for seeding the database with initial data.

## Overview

The application supports two seeding methods:

1. **SQL Dumps (Recommended)** - Snapshot-based seeding using SQL files
2. **CSV/JSON Scripts (Legacy)** - Programmatic seeding from CSV and JSON files

The migration system automatically prefers SQL dumps if available, falling back to CSV/JSON scripts if needed.

## SQL Dump-Based Seeding (Recommended)

### Why SQL Dumps?

- **Robust**: No normalization or data transformation needed
- **Complete**: Captures exact database state including all relationships
- **Fast**: Direct SQL import is faster than programmatic insertion
- **Idempotent**: Won't duplicate data on repeated runs
- **Easy to regenerate**: One command creates fresh dumps from your current database

### How It Works

When you run `docker compose up` or `go run ./cmd/migrate -seed`, the system:

1. Checks if SQL dump files exist in `backend/migrations/seed/sql-dumps/`
2. If dumps exist and database is empty, imports them in dependency order
3. If dumps don't exist, falls back to CSV/JSON-based seeding

### Generating SQL Dumps

When you have a database state you want to use as seed data:

```bash
# From project root
./scripts/dump-seed-data.sh
```

This creates SQL files in `backend/migrations/seed/sql-dumps/`:
- `schools.sql` - All schools/teams
- `tournaments.sql` - Tournament metadata
- `tournament_teams.sql` - Teams in each tournament
- `users.sql` - Seed users (admin, test users)
- `calcuttas.sql` - Calcutta pool data
- `calcutta_rounds.sql` - Round configurations
- `calcutta_entries.sql` - Player entries
- `calcutta_entry_teams.sql` - Team ownership

### When to Regenerate Dumps

Regenerate your SQL dumps when:
- You've added new seed data to your database
- You've fixed data issues and want to capture the clean state
- You've added new tournaments or calcuttas
- You want to update the default seed data for new developers

### Workflow Example

```bash
# 1. Set up your database with the data you want
docker compose up
# ... manually add/fix data via the app or SQL ...

# 2. Create a snapshot
./scripts/dump-seed-data.sh

# 3. Commit the dumps
git add backend/migrations/seed/sql-dumps/*.sql
git commit -m "Update seed data with 2025 tournament"

# 4. New developers automatically get this data
docker compose down -v
docker compose up  # Fresh database with your seed data!
```

## CSV/JSON-Based Seeding (Legacy)

This is the original seeding approach that builds data from CSV and JSON files.

### Available Scripts

Located in `backend/cmd/`:

- **seed-schools** - Loads schools from CSV
- **seed-tournaments** - Loads tournaments from JSON
- **seed-calcuttas** - Loads calcutta data from CSV

### Data Files

Located in `backend/migrations/seed/`:

- `schools/active_d1_teams.csv` - List of Division 1 schools
- `schools/schools.json` - School names in JSON format
- `calcuttas/*.csv` - Historical calcutta data by year
- `tournaments/*.sql` - Tournament seed SQL

### Running Manual Seeds

```bash
cd backend

# Seed schools
go run ./cmd/seed-schools

# Seed tournaments
go run ./cmd/seed-tournaments

# Seed calcuttas
go run ./cmd/seed-calcuttas
```

### When to Use CSV/JSON Scripts

- Building seed data from scratch
- You don't have a complete database to snapshot
- You need to transform or normalize data during import
- You're maintaining the CSV files as the source of truth

## Comparison

| Feature | SQL Dumps | CSV/JSON Scripts |
|---------|-----------|------------------|
| Setup complexity | Low | High |
| Maintenance | Easy (one command) | Manual (update CSVs) |
| Speed | Fast | Slower |
| Data integrity | Guaranteed | Requires validation |
| Normalization | Not needed | Required |
| Idempotent | Yes | Depends on script |
| Best for | Snapshots of working data | Building from scratch |

## Troubleshooting

### SQL Dumps Not Loading

Check that:
1. Dumps exist: `ls backend/migrations/seed/sql-dumps/*.sql`
2. Database is empty (dumps only load into empty databases)
3. Files aren't empty: `cat backend/migrations/seed/sql-dumps/schools.sql`

To force regeneration:
```bash
docker compose down -v  # Remove database
./scripts/dump-seed-data.sh  # Regenerate dumps
docker compose up  # Fresh start
```

### CSV Scripts Failing

Common issues:
- File paths incorrect (check working directory)
- Data normalization issues (check school name mappings)
- Foreign key violations (check data order)

See individual script files for detailed error messages.

## Recommendations

1. **For new projects**: Use CSV/JSON scripts initially to build seed data
2. **Once stable**: Generate SQL dumps and commit them
3. **For ongoing development**: Regenerate dumps periodically as data evolves
4. **Keep both**: Maintain CSV files as documentation, use SQL dumps for seeding
