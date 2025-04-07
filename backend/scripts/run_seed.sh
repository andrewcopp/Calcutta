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

# Function to run a SQL file
run_sql_file() {
    local file="$1"
    echo "Running $file..."
    echo "Full path: $(pwd)/$file"
    if [ ! -f "$file" ]; then
        echo "Error: File $file not found!"
        exit 1
    fi
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -v ON_ERROR_STOP=1 -f "$file"
    echo "Done running $file"
    echo "----------------------------------------"
}

# Run seed files in order
echo "Starting database seeding..."

# 1. Users (needed for calcutta ownership)
run_sql_file "migrations/seed/users/20240407_users.sql"

# 2. Schools (needed for tournament teams)
run_sql_file "migrations/seed/schools/20240407_schools.sql"

# 3. Tournaments (needed for teams and games)
run_sql_file "migrations/seed/tournaments/20240407_tournaments.sql"

# 4. Tournament teams and games
run_sql_file "migrations/seed/tournaments/20240407_tournament_data.sql"

# 5. Calcuttas, entries, and teams
run_sql_file "migrations/seed/calcuttas/20240407_calcutta_data.sql"

echo "Database seeding completed successfully!" 