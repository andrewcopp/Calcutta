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

# Run the seed-schools command
echo "Running seed-schools..."
cd "$(dirname "$0")/../cmd/seed-schools"
go run main.go

# Check if the command was successful
if [ $? -eq 0 ]; then
  echo "Successfully seeded schools"
else
  echo "Failed to seed schools"
  exit 1
fi 