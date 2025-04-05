#!/bin/bash

# Load environment variables from .env file
if [ -f ../.env ]; then
  export $(grep -v '^#' ../.env | xargs)
fi

# Override DB_HOST to use localhost
export DB_HOST=localhost

# Construct DATABASE_URL with SSL disabled
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Run the server
echo "Starting server..."
go run cmd/server/main.go 