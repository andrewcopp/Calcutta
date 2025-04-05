#!/bin/bash

# Load environment variables from .env file
if [ -f ../.env ]; then
  export $(grep -v '^#' ../.env | xargs)
fi

# Override DB_HOST to use localhost
export DB_HOST=localhost

# Construct DATABASE_URL with SSL disabled
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Run migrations
echo "Running database migrations..."
migrate -path internal/db/migrations -database "${DATABASE_URL}" up

if [ $? -eq 0 ]; then
  echo "Migrations completed successfully"
else
  echo "Migration failed"
  exit 1
fi 