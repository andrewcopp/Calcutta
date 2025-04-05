#!/bin/bash

# Load environment variables from .env file
if [ -f ../../.env ]; then
  export $(grep -v '^#' ../../.env | xargs)
fi

# Override DB_HOST to use localhost
export DB_HOST=localhost

# Construct DATABASE_URL
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Run the test
go run ../../cmd/testdb/main.go 