#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Set default database URL if not provided
if [ -z "$DATABASE_URL" ]; then
  export DATABASE_URL="postgres://postgres:postgres@localhost:5432/calcutta?sslmode=disable"
  echo "Using default DATABASE_URL: $DATABASE_URL"
fi

# Run the testdb command
echo "Testing database connection..."
cd "$(dirname "$0")/../cmd/testdb"
go run main.go

# Check if the command was successful
if [ $? -eq 0 ]; then
  echo "Database connection test successful"
else
  echo "Database connection test failed"
  exit 1
fi 