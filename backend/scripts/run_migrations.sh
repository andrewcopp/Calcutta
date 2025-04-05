#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Set default database URL if not provided
if [ -z "$DATABASE_URL" ]; then
  export DATABASE_URL="postgres://calcutta:calcutta@localhost:5432/calcutta?sslmode=disable"
  echo "Using default DATABASE_URL: $DATABASE_URL"
fi

# Run the migrate command
echo "Running migrations..."
cd "$(dirname "$0")/.."
go run cmd/migrate/main.go "$@"

# Check if the command was successful
if [ $? -eq 0 ]; then
  echo "Successfully ran migrations"
else
  echo "Failed to run migrations"
  exit 1
fi 