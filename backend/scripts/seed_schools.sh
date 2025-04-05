#!/bin/bash

# Load environment variables from .env file
if [ -f ../../.env ]; then
  export $(grep -v '^#' ../../.env | xargs)
fi

# Override DB_HOST to use localhost
export DB_HOST=localhost

# Construct DATABASE_URL with SSL disabled
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Run the seed script
echo "Seeding schools from master JSON file..."
go run ./seed/schools/seed_schools.go

if [ $? -eq 0 ]; then
  echo "Schools seeded successfully"
else
  echo "Failed to seed schools"
  exit 1
fi 