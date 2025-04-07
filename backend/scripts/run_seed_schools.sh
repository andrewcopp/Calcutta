#!/bin/bash

# Load environment variables from .env file
if [ -f ../../.env ]; then
  export $(grep -v '^#' ../../.env | xargs)
fi

# Override DB_HOST to use localhost
export DB_HOST=localhost

# Construct DATABASE_URL with SSL disabled
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Change to the backend directory
cd "$(dirname "$0")/.."

# Run the seed script
echo "Seeding schools from active_d1_teams.csv..."
go run ./cmd/seed-schools/main.go

if [ $? -eq 0 ]; then
  echo "Schools seeded successfully"
else
  echo "Failed to seed schools"
  exit 1
fi 