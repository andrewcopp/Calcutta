#!/bin/bash

# Script to dump seed data from the database into SQL files
# This creates a snapshot of the current database state for seeding

set -e

# Print debug info
echo "Current directory: $(pwd)"
echo "Script location: $(dirname "$0")"

# Load environment variables from .env file
if [ -f .env ]; then
    source .env
fi

# Default database connection values
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-calcutta}
DB_USER=${DB_USER:-calcutta}
DB_PASSWORD=${DB_PASSWORD:-calcutta}

# Force localhost for database connection (override Docker service name)
DB_HOST=localhost

# Set PGPASSWORD for pg_dump
export PGPASSWORD=$DB_PASSWORD

# Create output directory
OUTPUT_DIR="backend/migrations/seed/sql-dumps"
mkdir -p $OUTPUT_DIR

echo ""
echo "Using database connection:"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""
echo "Output directory: $OUTPUT_DIR"
echo ""

# Check if Docker is running and db container exists
if ! docker ps --format '{{.Names}}' | grep -q "^calcutta-db-1$"; then
    echo "Error: Docker database container 'calcutta-db-1' is not running."
    echo "Please start it with: docker compose up -d db"
    exit 1
fi

# Function to dump a table using Docker's pg_dump
dump_table() {
    local table=$1
    local output_file=$2
    local temp_file="${output_file}.tmp"
    
    echo "Dumping $table..."
    docker exec calcutta-db-1 pg_dump \
        -U $DB_USER \
        -d $DB_NAME \
        --data-only \
        --table=$table \
        --inserts \
        --no-owner \
        --no-privileges > $temp_file
    
    if [ $? -ne 0 ]; then
        echo "Error dumping $table"
        rm -f $temp_file
        exit 1
    fi
    
    # Remove psql meta-commands and other non-SQL lines that cause issues
    # Keep only the actual SQL statements (SET commands and INSERT statements)
    sed '/^\\restrict/d; /^-- Dumped from database/d; /^-- Dumped by pg_dump/d; /^--$/d' $temp_file > $output_file
    
    rm -f $temp_file
}

# Dump each table
dump_table "schools" "$OUTPUT_DIR/schools.sql"
dump_table "tournaments" "$OUTPUT_DIR/tournaments.sql"
dump_table "tournament_teams" "$OUTPUT_DIR/tournament_teams.sql"
dump_table "users" "$OUTPUT_DIR/users.sql"
dump_table "calcuttas" "$OUTPUT_DIR/calcuttas.sql"
dump_table "calcutta_rounds" "$OUTPUT_DIR/calcutta_rounds.sql"
dump_table "calcutta_entries" "$OUTPUT_DIR/calcutta_entries.sql"
dump_table "calcutta_entry_teams" "$OUTPUT_DIR/calcutta_entry_teams.sql"

# Create a master file that imports all dumps in the correct order
echo "Creating master import file..."
cat > $OUTPUT_DIR/00_import_all.sql << 'EOF'
-- Master seed data import file
-- This file imports all seed data in the correct order to maintain referential integrity
--
-- Usage: This file is automatically used by the seed migration process
-- To manually import: psql -d calcutta -f 00_import_all.sql

-- Disable triggers during import to avoid constraint issues
SET session_replication_role = replica;

-- Import in dependency order
\i schools.sql
\i tournaments.sql
\i tournament_teams.sql
\i users.sql
\i calcuttas.sql
\i calcutta_rounds.sql
\i calcutta_entries.sql
\i calcutta_entry_teams.sql

-- Re-enable triggers
SET session_replication_role = DEFAULT;
EOF

# Add timestamp to each file
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
for file in $OUTPUT_DIR/*.sql; do
    sed -i.bak "1i\\
-- Generated on: $TIMESTAMP\\
-- This file contains seed data for the Calcutta application\\
" "$file"
    rm "${file}.bak"
done

echo ""
echo "✅ Seed data dump complete!"
echo "Files created in: $OUTPUT_DIR"
echo ""
echo "To use these dumps for seeding:"
echo "  1. The dumps are automatically used by 'docker compose up'"
echo "  2. Or run manually: go run ./cmd/migrate -seed"
echo ""
echo "To create a fresh dump in the future:"
echo "  ./scripts/dump-seed-data.sh"
