# SQL Dump-Based Seed Data

This directory contains SQL dumps of the database seed data. This approach is more robust than CSV-based seeding because:

1. **No normalization needed** - Data is already in the correct format
2. **Referential integrity** - Foreign keys are preserved
3. **Complete snapshots** - Captures the exact state of your database
4. **Easy to regenerate** - Run the dump script whenever you need a fresh snapshot

## Files

- `00_import_all.sql` - Master import file that loads all dumps in the correct order
- `schools.sql` - School/team data
- `tournaments.sql` - Tournament metadata
- `tournament_teams.sql` - Teams participating in tournaments
- `users.sql` - Seed users (admin, test users, etc.)
- `calcuttas.sql` - Calcutta pool data
- `calcutta_rounds.sql` - Round configuration for each Calcutta
- `calcutta_entries.sql` - Player entries in Calcuttas
- `calcutta_entry_teams.sql` - Teams owned by each entry

## Usage

### Automatic (Recommended)

When you run `docker compose up`, the seed data is automatically loaded from these SQL dumps.

### Manual

```bash
# From the backend directory
go run ./cmd/migrate -seed
```

## Generating Fresh Dumps

When you have new data in your database that you want to use as seed data:

```bash
# From the project root
./scripts/dump-seed-data.sh
```

This will:
1. Connect to your database using `DATABASE_URL`
2. Export all seed tables to SQL files
3. Create a master import file with the correct order
4. Add timestamps to each file

## Migration from CSV

The CSV-based seeding scripts are still available in the parent directory. You can use either approach:

- **SQL dumps** (this directory) - Use for complete, tested database snapshots
- **CSV scripts** (`../calcuttas/`, `../schools/`, etc.) - Use for building seed data from scratch

The migration system will prefer SQL dumps if they exist, falling back to CSV-based seeding if needed.
