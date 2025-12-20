#!/bin/bash

# Script to clean existing SQL dump files by removing psql meta-commands

set -e

DUMPS_DIR="backend/migrations/seed/sql-dumps"

echo "Cleaning SQL dump files in $DUMPS_DIR..."

for file in $DUMPS_DIR/*.sql; do
    if [ -f "$file" ] && [ "$(basename "$file")" != "00_import_all.sql" ]; then
        echo "Cleaning $(basename "$file")..."
        
        # Create temp file
        temp_file="${file}.tmp"
        
        # Remove problematic lines
        sed '/^\\restrict/d' "$file" | \
        sed '/^-- Dumped from database/d' | \
        sed '/^-- Dumped by pg_dump/d' > "$temp_file"
        
        # Replace original with cleaned version
        mv "$temp_file" "$file"
    fi
done

echo "✅ SQL dump files cleaned successfully!"
