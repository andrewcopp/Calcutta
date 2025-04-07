#!/bin/bash

# Exit on error
set -e

# Print debug info
echo "Current directory: $(pwd)"
echo "Script location: $(dirname "$0")"

# Ensure we're in the backend directory
cd "$(dirname "$0")/.."
echo "Changed to directory: $(pwd)"

# Force localhost for database connection
DB_HOST=localhost

# Load environment variables
if [ -f .env ]; then
    source .env
fi

# Default database connection
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-calcutta}
DB_USER=${DB_USER:-calcutta}
DB_PASSWORD=${DB_PASSWORD:-calcutta}

# Print database connection info
echo "Using database connection:"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"

# Function to dump a table to a SQL file
dump_table() {
    local table=$1
    local file=$2
    echo "Dumping $table to $file..."
    PGPASSWORD=$DB_PASSWORD /opt/homebrew/opt/postgresql@16/bin/pg_dump \
        -h $DB_HOST \
        -p $DB_PORT \
        -U $DB_USER \
        -d $DB_NAME \
        --data-only \
        --table=public.$table \
        > "$file"
    echo "Done dumping $table"
    echo "----------------------------------------"
}

# Create timestamp for filenames
TIMESTAMP=$(date +%Y%m%d)

# Update seed files
echo "Starting database seed updates..."

# 1. Users
dump_table "users" "migrations/seed/users/${TIMESTAMP}_users.sql"

# 2. Schools
dump_table "schools" "migrations/seed/schools/${TIMESTAMP}_schools.sql"

# 3. Tournaments
dump_table "tournaments" "migrations/seed/tournaments/${TIMESTAMP}_tournaments.sql"

# 4. Tournament Teams
dump_table "tournament_teams" "migrations/seed/tournaments/${TIMESTAMP}_tournament_teams.sql"

# 5. Calcuttas and related tables
dump_table "calcuttas" "migrations/seed/calcuttas/${TIMESTAMP}_calcuttas.sql"
dump_table "calcutta_entries" "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_entries.sql"
dump_table "calcutta_entry_teams" "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_entry_teams.sql"
dump_table "calcutta_rounds" "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_rounds.sql"

echo "Database seed updates completed successfully!" 