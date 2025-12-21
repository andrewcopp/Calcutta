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

# Create a log directory if it doesn't exist
mkdir -p logs

# Function to run a SQL file
run_sql_file() {
    local file="$1"
    local log_file="logs/$(basename "$file").log"
    echo "Running $file..."
    echo "Full path: $(pwd)/$file"
    if [ ! -f "$file" ]; then
        echo "Error: File $file not found!"
        exit 1
    fi
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -v ON_ERROR_STOP=1 -q -f "$file" > "$log_file" 2>&1
    if [ $? -eq 0 ]; then
        echo "Done running $file"
    else
        echo "Error running $file. Check $log_file for details."
        exit 1
    fi
    echo "----------------------------------------"
}

# Get the latest seed files
TIMESTAMP="$1"
if [ -z "$TIMESTAMP" ]; then
    TIMESTAMP=$(ls -1 migrations/seed/users/*_users.sql 2>/dev/null | sed -E 's#.*\/([0-9]{8})_users\.sql#\1#' | sort | tail -n 1)
fi
if [ -z "$TIMESTAMP" ]; then
    echo "Error: Could not determine seed timestamp (no migrations/seed/users/*_users.sql files found)"
    exit 1
fi

echo "Using seed timestamp: $TIMESTAMP"

# Run seed files in order
echo "Starting database seeding..."

# 1. Users (needed for calcutta ownership)
run_sql_file "migrations/seed/users/${TIMESTAMP}_users.sql"

# 2. Schools (needed for tournament teams)
run_sql_file "migrations/seed/schools/${TIMESTAMP}_schools.sql"

# 3. Tournaments (needed for teams and games)
run_sql_file "migrations/seed/tournaments/${TIMESTAMP}_tournaments.sql"

# 4. Tournament teams
run_sql_file "migrations/seed/tournaments/${TIMESTAMP}_tournament_teams.sql"

# 5. Calcuttas and related tables
run_sql_file "migrations/seed/calcuttas/${TIMESTAMP}_calcuttas.sql"
run_sql_file "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_entries.sql"
run_sql_file "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_entry_teams.sql"
run_sql_file "migrations/seed/calcuttas/${TIMESTAMP}_calcutta_rounds.sql"

echo "Database seeding completed successfully!" 